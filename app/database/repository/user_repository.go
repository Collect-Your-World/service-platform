package repository

import (
	"backend/service-platform/app/database/constant/user"
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/internal/runtime"
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	Insert(ctx context.Context, user *entity.User) (*entity.User, error)
	Update(ctx context.Context, user entity.User) (*entity.User, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.User, error)
	FindActiveUserByUsername(ctx context.Context, username string) (*entity.User, error)
	UpdateLastLoginAt(ctx context.Context, userID uuid.UUID) error
	UpdateEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error
	UpdatePhoneVerified(ctx context.Context, userID uuid.UUID, verified bool) error
}

type DefaultUserRepository struct {
	res runtime.Resource
}

func NewUserRepository(res runtime.Resource) UserRepository {
	return &DefaultUserRepository{res: res}
}

func (r DefaultUserRepository) Insert(ctx context.Context, user *entity.User) (*entity.User, error) {
	err := r.res.DB.
		NewInsert().
		Model(user).
		Returning("*").
		Scan(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r DefaultUserRepository) Update(ctx context.Context, user entity.User) (*entity.User, error) {
	var u entity.User
	err := r.res.DB.
		NewUpdate().
		Model(&user).
		WherePK().
		Where("deleted_at IS NULL").
		Returning("*").
		Scan(ctx, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r DefaultUserRepository) DeleteByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var u entity.User
	err := r.res.DB.
		NewUpdate().
		Model(&u).
		Set("deleted_at", time.Now()).
		WherePK().
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Returning("*").
		Scan(ctx, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r DefaultUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	u := new(entity.User)
	err := r.res.DB.
		ReplicaNewSelect().
		Model(u).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r DefaultUserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	u := new(entity.User)
	err := r.res.DB.
		ReplicaNewSelect().
		Model(u).
		Where("username = ?", username).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r DefaultUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	u := new(entity.User)
	err := r.res.DB.
		ReplicaNewSelect().
		Model(u).
		Where("email = ?", email).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Removed principal lookups; users are referenced directly by ID

func (r DefaultUserRepository) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.User, error) {
	u := new(entity.User)
	err := r.res.DB.
		ReplicaNewSelect().
		Model(u).
		Where("phone_number = ?", phoneNumber).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r DefaultUserRepository) FindActiveUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	u := new(entity.User)
	err := r.res.DB.
		ReplicaNewSelect().
		Model(u).
		Where("username = ?", username).
		Where("status = ?", user.Verified).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r DefaultUserRepository) UpdateLastLoginAt(ctx context.Context, userID uuid.UUID) error {
	_, err := r.res.DB.
		NewUpdate().
		Model((*entity.User)(nil)).
		Set("last_login_at = ?", time.Now()).
		Where("id = ?", userID).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}

func (r DefaultUserRepository) UpdateEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	_, err := r.res.DB.
		NewUpdate().
		Model((*entity.User)(nil)).
		Set("email_verified = ?", verified).
		Where("id = ?", userID).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}

func (r DefaultUserRepository) UpdatePhoneVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	_, err := r.res.DB.
		NewUpdate().
		Model((*entity.User)(nil)).
		Set("phone_verified = ?", verified).
		Where("id = ?", userID).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}
