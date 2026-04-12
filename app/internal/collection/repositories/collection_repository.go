package repository

import (
	collectionconst "backend/service-platform/app/internal/collection/constants/collection"
	"backend/service-platform/app/internal/collection/entities"
	"backend/service-platform/app/internal/platform/runtime"
	"backend/service-platform/app/internal/user/constants/currency"
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CollectionFilter struct {
	IDs              []uuid.UUID
	Names            []string
	Slugs            []string
	Types            []collectionconst.Type
	RewardCurrencies []currency.Currency
	RewardAmounts    []int64
	MinRewardAmount  *int64
	MaxRewardAmount  *int64
	IsEnabled        *bool
	IncludeDeleted   bool
	Limit            int
	Offset           int
}

type CollectionRepository interface {
	Create(ctx context.Context, collection *entity.Collection) (*entity.Collection, error)
	Update(ctx context.Context, collection *entity.Collection) (*entity.Collection, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Collection, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Collection, error)
	List(ctx context.Context, filter CollectionFilter) ([]entity.Collection, error)
	SoftDelete(ctx context.Context, exec bun.IDB, id uuid.UUID) (int64, error)
}

type DefaultCollectionRepository struct {
	res runtime.Resource
}

func NewCollectionRepository(res runtime.Resource) CollectionRepository {
	return &DefaultCollectionRepository{res: res}
}

func (r *DefaultCollectionRepository) Create(ctx context.Context, collection *entity.Collection) (*entity.Collection, error) {
	err := r.res.DB.NewInsert().Model(collection).Returning("*").Scan(ctx, collection)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func (r *DefaultCollectionRepository) Update(ctx context.Context, collection *entity.Collection) (*entity.Collection, error) {
	var updated entity.Collection
	err := r.res.DB.NewUpdate().
		Model(collection).
		WherePK().
		Where("deleted_at IS NULL").
		Returning("*").
		Scan(ctx, &updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *DefaultCollectionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Collection, error) {
	collection := new(entity.Collection)
	err := r.res.DB.ReplicaNewSelect().
		Model(collection).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func (r *DefaultCollectionRepository) FindBySlug(ctx context.Context, slug string) (*entity.Collection, error) {
	collection := new(entity.Collection)
	err := r.res.DB.ReplicaNewSelect().
		Model(collection).
		Where("slug = ?", slug).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

func (r *DefaultCollectionRepository) List(ctx context.Context, filter CollectionFilter) ([]entity.Collection, error) {
	var collections []entity.Collection
	query := r.res.DB.ReplicaNewSelect().
		Model(&collections).
		OrderExpr("created_at DESC")

	if filter.IncludeDeleted {
		query = query.WhereAllWithDeleted()
	} else {
		query = query.Where("deleted_at IS NULL")
	}

	if len(filter.IDs) > 0 {
		query = query.Where("id IN (?)", bun.In(filter.IDs))
	}
	if len(filter.Names) > 0 {
		query = query.Where("name IN (?)", bun.In(filter.Names))
	}
	if len(filter.Slugs) > 0 {
		query = query.Where("slug IN (?)", bun.In(filter.Slugs))
	}
	if len(filter.Types) > 0 {
		query = query.Where("collection_type IN (?)", bun.In(filter.Types))
	}
	if len(filter.RewardCurrencies) > 0 {
		query = query.Where("reward_currency IN (?)", bun.In(filter.RewardCurrencies))
	}
	if len(filter.RewardAmounts) > 0 {
		query = query.Where("reward_amount IN (?)", bun.In(filter.RewardAmounts))
	}
	if filter.MinRewardAmount != nil {
		query = query.Where("reward_amount >= ?", *filter.MinRewardAmount)
	}
	if filter.MaxRewardAmount != nil {
		query = query.Where("reward_amount <= ?", *filter.MaxRewardAmount)
	}
	if filter.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *filter.IsEnabled)
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
	return collections, nil
}

func (r *DefaultCollectionRepository) SoftDelete(ctx context.Context, exec bun.IDB, id uuid.UUID) (int64, error) {
	db := r.execDB(exec)
	res, err := db.NewUpdate().
		Model((*entity.Collection)(nil)).
		Set("deleted_at = NOW()").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

func (r *DefaultCollectionRepository) execDB(exec bun.IDB) bun.IDB {
	if exec != nil {
		return exec
	}
	return r.res.DB
}
