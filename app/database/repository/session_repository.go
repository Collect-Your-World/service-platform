package repository

import (
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/internal/runtime"
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionRepository interface {
	Insert(ctx context.Context, session *entity.Session) (*entity.Session, error)
	FindByToken(ctx context.Context, token string) (*entity.Session, error)
	RevokeByToken(ctx context.Context, token string) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteByID(ctx context.Context, id uuid.UUID) (*entity.Session, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Session, error)
}

type DefaultSessionRepository struct {
	res runtime.Resource
}

func NewSessionRepository(res runtime.Resource) SessionRepository {
	return &DefaultSessionRepository{res: res}
}

func (r DefaultSessionRepository) Insert(ctx context.Context, session *entity.Session) (*entity.Session, error) {
	// Revoke existing token if duplicated
	_ = r.RevokeByToken(ctx, session.Token)
	// Soft-revoke existing sessions for the same user
	_ = r.RevokeByUserID(ctx, session.UserID)

	err := r.res.DB.NewInsert().Model(session).Returning("*").Scan(ctx, session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r DefaultSessionRepository) FindByToken(ctx context.Context, token string) (*entity.Session, error) {
	var session entity.Session
	err := r.res.DB.ReplicaNewSelect().Model(&session).Where("token = ?", token).Where("deleted_at IS NULL").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r DefaultSessionRepository) RevokeByToken(ctx context.Context, token string) error {
	_, err := r.res.DB.NewUpdate().Model((*entity.Session)(nil)).Set("revoked = ?", true).Set("deleted_at = ?", time.Now()).Where("token = ?", token).Where("deleted_at IS NULL").Exec(ctx)
	return err
}

func (r DefaultSessionRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.res.DB.NewUpdate().Model((*entity.Session)(nil)).Set("revoked = ?", true).Set("deleted_at = ?", time.Now()).Where("user_id = ?", userID).Where("deleted_at IS NULL").Exec(ctx)
	return err
}

func (r DefaultSessionRepository) DeleteByID(ctx context.Context, id uuid.UUID) (*entity.Session, error) {
	var session entity.Session
	err := r.res.DB.NewUpdate().Model(&session).Set("deleted_at", time.Now()).WherePK().Where("id = ?", id).Where("deleted_at IS NULL").Returning("*").Scan(ctx, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r DefaultSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Session, error) {
	var sessions []entity.Session
	err := r.res.DB.NewUpdate().Model(&sessions).Set("deleted_at = ?", time.Now()).Where("user_id = ?", userID).Where("deleted_at IS NULL").Returning("*").Scan(ctx, &sessions)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}
