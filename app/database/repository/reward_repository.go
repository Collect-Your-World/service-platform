package repository

import (
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/internal/runtime"
	"context"
	"time"

	"github.com/google/uuid"
)

type RewardRepository interface {
	Insert(ctx context.Context, reward *entity.Reward) (*entity.Reward, error)
	Update(ctx context.Context, reward entity.Reward) (*entity.Reward, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (*entity.Reward, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Reward, error)
	FindByCode(ctx context.Context, code string) (*entity.Reward, error)
	ListAll(ctx context.Context) ([]entity.Reward, error)
}

type DefaultRewardRepository struct {
	res runtime.Resource
}

func NewRewardRepository(res runtime.Resource) RewardRepository {
	return &DefaultRewardRepository{res: res}
}

func (r DefaultRewardRepository) Insert(ctx context.Context, reward *entity.Reward) (*entity.Reward, error) {
	err := r.res.DB.
		NewInsert().
		Model(reward).
		Returning("*").
		Scan(ctx, reward)
	if err != nil {
		return nil, err
	}
	return reward, nil
}

func (r DefaultRewardRepository) Update(ctx context.Context, reward entity.Reward) (*entity.Reward, error) {
	var updated entity.Reward
	err := r.res.DB.
		NewUpdate().
		Model(&reward).
		WherePK().
		Where("deleted_at IS NULL").
		Returning("*").
		Scan(ctx, &updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r DefaultRewardRepository) DeleteByID(ctx context.Context, id uuid.UUID) (*entity.Reward, error) {
	var deleted entity.Reward
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

func (r DefaultRewardRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Reward, error) {
	reward := new(entity.Reward)
	err := r.res.DB.
		ReplicaNewSelect().
		Model(reward).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return reward, nil
}

func (r DefaultRewardRepository) FindByCode(ctx context.Context, code string) (*entity.Reward, error) {
	reward := new(entity.Reward)
	err := r.res.DB.
		ReplicaNewSelect().
		Model(reward).
		Where("code = ?", code).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return reward, nil
}

func (r DefaultRewardRepository) ListAll(ctx context.Context) ([]entity.Reward, error) {
	var rewards []entity.Reward
	err := r.res.DB.
		ReplicaNewSelect().
		Model(&rewards).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return rewards, nil
}
