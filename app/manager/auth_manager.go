package manager

import (
	"backend/service-platform/app/api/client/request"
	"backend/service-platform/app/api/client/response"
	"backend/service-platform/app/database/constant/role"
	userstatus "backend/service-platform/app/database/constant/user"
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/database/repository"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/pkg/bcrypt"
	"backend/service-platform/app/pkg/jwt"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	ErrUsernameAlreadyExisted = "username already exists"
	ErrEmailAlreadyExists     = "email already exists"
	ErrInvalidCredentials     = "invalid credentials"
	ErrInvalidRefreshToken    = "invalid refresh token"
	ErrRefreshTokenExpired    = "refresh token has expired"
	ErrRefreshTokenRevoked    = "refresh token has been revoked"
)

type AuthManager interface {
	Logout(ctx context.Context, request request.LogoutRequest) error
	Login(ctx context.Context, request request.AuthUserRequest) (*response.AuthResponse, error)
	RefreshToken(ctx context.Context, request request.RefreshTokenRequest) (*response.AuthResponse, error)
	Register(ctx context.Context, request request.RegisterRequest) error
}

type DefaultAuthManager struct {
	logger       *zap.Logger
	res          runtime.Resource
	hasher       bcrypt.Hasher
	jwtManager   jwt.Jwt
	repositories *repository.Repositories
}

func NewAuthManager(
	res runtime.Resource,
	hasher bcrypt.Hasher,
	jwtManager jwt.Jwt,
	repositories *repository.Repositories,
) AuthManager {
	return &DefaultAuthManager{
		res:          res,
		logger:       res.Logger,
		hasher:       hasher,
		jwtManager:   jwtManager,
		repositories: repositories,
	}
}

func (d *DefaultAuthManager) Logout(ctx context.Context, request request.LogoutRequest) error {
	claims, err := d.jwtManager.ValidateToken(request.RefreshToken)
	if err != nil || claims.RefreshTokenBase64 == nil || *claims.RefreshTokenBase64 == "" {
		return errors.New(ErrInvalidRefreshToken)
	}
	h := sha256.Sum256([]byte(*claims.RefreshTokenBase64))
	hashed := hex.EncodeToString(h[:])
	if err := d.repositories.SessionRepository.RevokeByToken(ctx, hashed); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}

func (d *DefaultAuthManager) Register(ctx context.Context, request request.RegisterRequest) error {
	_, err := d.repositories.UserRepository.FindByEmail(ctx, request.Email)
	if err == nil {
		return errors.New(ErrEmailAlreadyExists)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Use email as username for now
	_, err = d.repositories.UserRepository.FindByUsername(ctx, request.Email)
	if err == nil {
		return errors.New(ErrUsernameAlreadyExisted)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	hashed, err := d.hasher.HashPassword(request.Password)
	if err != nil {
		return err
	}

	user := &entity.User{
		Username:      request.Email,
		Email:         &request.Email,
		Password:      hashed,
		Status:        userstatus.Unverified,
		Role:          role.User,
		EmailVerified: false,
		PhoneVerified: false,
	}

	_, err = d.repositories.UserRepository.Insert(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func (d *DefaultAuthManager) Login(ctx context.Context, request request.AuthUserRequest) (*response.AuthResponse, error) {
	u, err := d.repositories.UserRepository.FindByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Verify password
	valid, err := d.hasher.CheckPassword(request.Password, u.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to check password: %w", err)
	}
	if !valid {
		return nil, errors.New(ErrInvalidCredentials)
	}

	// Update last login timestamp
	if err := d.repositories.UserRepository.UpdateLastLoginAt(ctx, u.ID); err != nil {
		d.logger.Warn("failed to update last login timestamp", zap.Error(err))
	}

	// Update user's last login time for token generation
	now := time.Now()
	u.LastLoginAt = &now

	accessToken, err := d.generateUserAccessToken(ctx, u)
	if err != nil {
		return nil, err
	}
	refreshTokenString, err := d.createSession(ctx, u)
	if err != nil {
		return nil, err
	}

	// Get roles for response
	userRoles := []role.Role{u.Role}
	resp := d.createAuthResponse(&u.Username, &userRoles, accessToken.Token, refreshTokenString)
	return resp, nil
}

func (d *DefaultAuthManager) RefreshToken(
	ctx context.Context,
	request request.RefreshTokenRequest,
) (*response.AuthResponse, error) {
	// Validate session token and get session info
	session, err := d.validateSession(ctx, request.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Find user by session
	u, err := d.repositories.UserRepository.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	accessToken, err := d.generateUserAccessToken(ctx, u)
	if err != nil {
		return nil, err
	}

	// Get roles for response
	userRoles := []role.Role{u.Role}
	return d.createAuthResponse(&u.Username, &userRoles, accessToken.Token, request.RefreshToken), nil
}

// Me not required in simplified flow; controller reads claims

// removed principal-based refresh

// getUserRoles extracts roles for a user by user ID
// roles removed

// validateSession validates and retrieves session information
func (d *DefaultAuthManager) validateSession(ctx context.Context, token string) (*entity.Session, error) {
	claims, err := d.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, errors.New(ErrInvalidRefreshToken)
	}
	if claims.RefreshTokenBase64 == nil || *claims.RefreshTokenBase64 == "" {
		return nil, errors.New(ErrInvalidRefreshToken)
	}
	h := sha256.Sum256([]byte(*claims.RefreshTokenBase64))
	hashed := hex.EncodeToString(h[:])
	session, err := d.repositories.SessionRepository.FindByToken(ctx, hashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(ErrInvalidRefreshToken)
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}
	if session.Revoked {
		return nil, errors.New(ErrRefreshTokenRevoked)
	}
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		return nil, errors.New(ErrRefreshTokenExpired)
	}
	return session, nil
}

func (d *DefaultAuthManager) createSession(ctx context.Context, user *entity.User) (string, error) {
	roleStr := string(user.Role)
	refreshToken, err := d.jwtManager.GenerateRefreshToken(
		&user.ID,
		&user.Username,
		user.Email,
		user.PhoneNumber,
		&roleStr,
		&user.EmailVerified,
		&user.PhoneVerified,
		user.LastLoginAt,
	)
	if err != nil {
		return "", err
	}

	h := sha256.Sum256([]byte(refreshToken.TokenBase64))
	hashed := hex.EncodeToString(h[:])

	exp := time.Now().Add(d.res.Config.JwtConfig.RefreshExpiration)
	_, err = d.repositories.SessionRepository.Insert(ctx, &entity.Session{UserID: user.ID, Token: hashed, ExpiresAt: &exp})
	if err != nil {
		return "", err
	}
	return refreshToken.Token, nil
}

// createAuthResponse creates a standardized auth response
func (d *DefaultAuthManager) createAuthResponse(
	username *string,
	roles *[]role.Role,
	accessToken string,
	refreshToken string,
) *response.AuthResponse {
	return &response.AuthResponse{
		Username:     username,
		Roles:        roles,
		AccessToken:  accessToken,
		ExpiresIn:    d.jwtManager.GetExpirationTime(),
		TokenType:    jwt.TokenTypeBearer,
		RefreshToken: refreshToken,
	}
}

// generateUserAccessToken creates an access token for a user with full user information
func (d *DefaultAuthManager) generateUserAccessToken(ctx context.Context, user *entity.User) (*jwt.AccessToken, error) {
	roleStr := string(user.Role)
	accessToken, err := d.jwtManager.GenerateAccessToken(
		&user.ID,
		&user.Username,
		user.Email,
		user.PhoneNumber,
		&roleStr,
		&user.EmailVerified,
		&user.PhoneVerified,
		user.LastLoginAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	return accessToken, nil
}
