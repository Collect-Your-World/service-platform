package repository

import (
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/internal/runtime"
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRewardRepository interface {
	Insert(ctx context.Context, ur *entity.UserReward) (*entity.UserReward, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (*entity.UserReward, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.UserReward, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.UserReward, error)
	FindByRewardID(ctx context.Context, rewardID uuid.UUID) ([]entity.UserReward, error)
}

type DefaultUserRewardRepository struct {
	res runtime.Resource
}

func NewUserRewardRepository(res runtime.Resource) UserRewardRepository {
	return &DefaultUserRewardRepository{res: res}
}

func (r DefaultUserRewardRepository) Insert(ctx context.Context, ur *entity.UserReward) (*entity.UserReward, error) {
	err := r.res.DB.
		NewInsert().
		Model(ur).
		Returning("*").
		Scan(ctx, ur)
	if err != nil {
		return nil, err
	}
	return ur, nil
}

func (r DefaultUserRewardRepository) DeleteByID(ctx context.Context, id uuid.UUID) (*entity.UserReward, error) {
	var deleted entity.UserReward
	err := r.res.DB.
		NewUpdate().
		Model(&deleted).
		Set("deleted_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Returning("*").
		Scan(ctx, &deleted)
	if err != nil {
		return nil, err
	}
	return &deleted, nil
}

func (r DefaultUserRewardRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.UserReward, error) {
	ur := new(entity.UserReward)
	err := r.res.DB.
		ReplicaNewSelect().
		Model(ur).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return ur, nil
}

func (r DefaultUserRewardRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.UserReward, error) {
	var list []entity.UserReward
	err := r.res.DB.
		ReplicaNewSelect().
		Model(&list).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Order("granted_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r DefaultUserRewardRepository) FindByRewardID(ctx context.Context, rewardID uuid.UUID) ([]entity.UserReward, error) {
	var list []entity.UserReward
	err := r.res.DB.
		ReplicaNewSelect().
		Model(&list).
		Where("reward_id = ?", rewardID).
		Where("deleted_at IS NULL").
		Order("granted_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}
