package repository

import (
	"backend/service-platform/app/internal/collection/entities"
	"backend/service-platform/app/internal/platform/runtime"
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ItemRepository interface {
	Create(ctx context.Context, item *entity.Item) (*entity.Item, error)
	Update(ctx context.Context, item *entity.Item) (*entity.Item, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Item, error)
	SoftDeleteByIDs(ctx context.Context, exec bun.IDB, ids []uuid.UUID) (int64, error)
}

type DefaultItemRepository struct {
	res runtime.Resource
}

func NewItemRepository(res runtime.Resource) ItemRepository {
	return &DefaultItemRepository{res: res}
}

func (r *DefaultItemRepository) Create(ctx context.Context, item *entity.Item) (*entity.Item, error) {
	err := r.res.DB.NewInsert().Model(item).Returning("*").Scan(ctx, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *DefaultItemRepository) Update(ctx context.Context, item *entity.Item) (*entity.Item, error) {
	var updated entity.Item
	err := r.res.DB.NewUpdate().Model(item).WherePK().Where("deleted_at IS NULL").Returning("*").Scan(ctx, &updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *DefaultItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Item, error) {
	item := new(entity.Item)
	err := r.res.DB.ReplicaNewSelect().Model(item).Where("id = ?", id).Where("deleted_at IS NULL").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *DefaultItemRepository) SoftDeleteByIDs(ctx context.Context, exec bun.IDB, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	db := r.execDB(exec)
	res, err := db.NewUpdate().Model((*entity.Item)(nil)).Set("deleted_at = NOW()").Where("id IN (?)", bun.In(ids)).Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

func (r *DefaultItemRepository) execDB(exec bun.IDB) bun.IDB {
	if exec != nil {
		return exec
	}
	return r.res.DB
}
