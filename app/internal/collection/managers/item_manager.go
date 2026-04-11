package managers

import (
	"backend/service-platform/app/internal/collection/entities"
	collrepo "backend/service-platform/app/internal/collection/repositories"
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ItemManager interface {
	CreateItem(ctx context.Context, item *entity.Item) (*entity.Item, error)
	UpdateItem(ctx context.Context, item *entity.Item) (*entity.Item, error)
	GetItem(ctx context.Context, id uuid.UUID) (*entity.Item, error)
	ListItemsByCollectionIDs(ctx context.Context, collectionIDs []uuid.UUID, includeDeleted bool) ([]entity.Item, error)
	DeleteItems(ctx context.Context, ids []uuid.UUID) error
}

type DefaultItemManager struct {
	items    collrepo.ItemRepository
	txRunner TransactionRunner
}

func NewItemManager(txRunner TransactionRunner, items collrepo.ItemRepository) ItemManager {
	return &DefaultItemManager{
		items:    items,
		txRunner: txRunner,
	}
}

func (m *DefaultItemManager) CreateItem(ctx context.Context, item *entity.Item) (*entity.Item, error) {
	return m.items.Create(ctx, item)
}

func (m *DefaultItemManager) UpdateItem(ctx context.Context, item *entity.Item) (*entity.Item, error) {
	return m.items.Update(ctx, item)
}

func (m *DefaultItemManager) GetItem(ctx context.Context, id uuid.UUID) (*entity.Item, error) {
	return m.items.FindByID(ctx, id)
}

func (m *DefaultItemManager) ListItemsByCollectionIDs(ctx context.Context, collectionIDs []uuid.UUID, includeDeleted bool) ([]entity.Item, error) {
	return m.items.ListByCollectionIDs(ctx, collectionIDs, includeDeleted)
}

func (m *DefaultItemManager) DeleteItems(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return m.txRunner.RunInTx(ctx, &sql.TxOptions{}, func(txCtx context.Context, tx bun.Tx) error {
		_, err := m.items.SoftDeleteByIDs(txCtx, tx, ids)
		return err
	})
}
