package repository

import (
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/internal/runtime"
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type RarityConfigFilter struct {
	IDs            []uuid.UUID
	Codes          []string
	IncludeDeleted bool
	Limit          int
	Offset         int
}

type RarityConfigRepository interface {
	Create(ctx context.Context, rc *entity.RarityConfig) (*entity.RarityConfig, error)
	Update(ctx context.Context, rc *entity.RarityConfig) (*entity.RarityConfig, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.RarityConfig, error)
	FindByCode(ctx context.Context, code string) (*entity.RarityConfig, error)
	List(ctx context.Context, filter RarityConfigFilter) ([]entity.RarityConfig, error)
}

type DefaultRarityConfigRepository struct {
	res runtime.Resource
}

func NewRarityConfigRepository(res runtime.Resource) RarityConfigRepository {
	return &DefaultRarityConfigRepository{res: res}
}

func (r *DefaultRarityConfigRepository) Create(ctx context.Context, rc *entity.RarityConfig) (*entity.RarityConfig, error) {
	err := r.res.DB.NewInsert().Model(rc).Returning("*").Scan(ctx, rc)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func (r *DefaultRarityConfigRepository) Update(ctx context.Context, rc *entity.RarityConfig) (*entity.RarityConfig, error) {
	var updated entity.RarityConfig
	err := r.res.DB.NewUpdate().Model(rc).WherePK().Where("deleted_at IS NULL").Returning("*").Scan(ctx, &updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *DefaultRarityConfigRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.RarityConfig, error) {
	rc := new(entity.RarityConfig)
	err := r.res.DB.ReplicaNewSelect().Model(rc).Where("id = ?", id).Where("deleted_at IS NULL").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func (r *DefaultRarityConfigRepository) FindByCode(ctx context.Context, code string) (*entity.RarityConfig, error) {
	rc := new(entity.RarityConfig)
	err := r.res.DB.ReplicaNewSelect().Model(rc).Where("code = ?", code).Where("deleted_at IS NULL").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func (r *DefaultRarityConfigRepository) List(ctx context.Context, filter RarityConfigFilter) ([]entity.RarityConfig, error) {
	var items []entity.RarityConfig
	query := r.res.DB.ReplicaNewSelect().Model(&items).OrderExpr("rank ASC")
	if filter.IncludeDeleted {
		query = query.WhereAllWithDeleted()
	} else {
		query = query.Where("deleted_at IS NULL")
	}
	if len(filter.IDs) > 0 {
		query = query.Where("id IN (?)", bun.In(filter.IDs))
	}
	if len(filter.Codes) > 0 {
		query = query.Where("code IN (?)", bun.In(filter.Codes))
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}
