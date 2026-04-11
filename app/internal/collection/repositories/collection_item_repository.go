package repository

import (
	"backend/service-platform/app/internal/collection/entities"
	"backend/service-platform/app/internal/platform/runtime"
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CollectionItemRepository interface {
	Create(ctx context.Context, item *entity.CollectionItem) (*entity.CollectionItem, error)
	Update(ctx context.Context, item *entity.CollectionItem) (*entity.CollectionItem, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.CollectionItem, error)
	ListByCollectionIDs(ctx context.Context, collectionIDs []uuid.UUID, includeDeleted bool) ([]entity.CollectionItem, error)
	SoftDeleteByCollectionID(ctx context.Context, exec bun.IDB, collectionID uuid.UUID) (int64, error)
	SoftDeleteByCollectionIDs(ctx context.Context, exec bun.IDB, collectionIDs []uuid.UUID) (int64, error)
}

type DefaultCollectionItemRepository struct {
	res runtime.Resource
}

func NewCollectionItemRepository(res runtime.Resource) CollectionItemRepository {
	return &DefaultCollectionItemRepository{res: res}
}

func (r *DefaultCollectionItemRepository) Create(ctx context.Context, item *entity.CollectionItem) (*entity.CollectionItem, error) {
	err := r.res.DB.NewInsert().Model(item).Returning("*").Scan(ctx, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *DefaultCollectionItemRepository) Update(ctx context.Context, item *entity.CollectionItem) (*entity.CollectionItem, error) {
	var updated entity.CollectionItem
	err := r.res.DB.NewUpdate().
		Model(item).
		WherePK().
		Where("deleted_at IS NULL").
		Returning("*").
		Scan(ctx, &updated)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *DefaultCollectionItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.CollectionItem, error) {
	item := new(entity.CollectionItem)
	err := r.res.DB.ReplicaNewSelect().
		Model(item).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *DefaultCollectionItemRepository) ListByCollectionIDs(ctx context.Context, collectionIDs []uuid.UUID, includeDeleted bool) ([]entity.CollectionItem, error) {
	if len(collectionIDs) == 0 {
		return []entity.CollectionItem{}, nil
	}
	var items []entity.CollectionItem
	query := r.res.DB.ReplicaNewSelect().
		Model(&items).
		Where("collection_id IN (?)", bun.In(collectionIDs)).
		OrderExpr("created_at ASC")
	if includeDeleted {
		query = query.WhereAllWithDeleted()
	} else {
		query = query.Where("deleted_at IS NULL")
	}
	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DefaultCollectionItemRepository) SoftDeleteByCollectionID(ctx context.Context, exec bun.IDB, collectionID uuid.UUID) (int64, error) {
	return r.softDelete(ctx, exec, []uuid.UUID{collectionID})
}

func (r *DefaultCollectionItemRepository) SoftDeleteByCollectionIDs(ctx context.Context, exec bun.IDB, collectionIDs []uuid.UUID) (int64, error) {
	return r.softDelete(ctx, exec, collectionIDs)
}

func (r *DefaultCollectionItemRepository) softDelete(ctx context.Context, exec bun.IDB, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	db := r.execDB(exec)
	res, err := db.NewUpdate().
		Model((*entity.CollectionItem)(nil)).
		Set("deleted_at = NOW()").
		Where("collection_id IN (?)", bun.In(ids)).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

func (r *DefaultCollectionItemRepository) execDB(exec bun.IDB) bun.IDB {
	if exec != nil {
		return exec
	}
	return r.res.DB
}
