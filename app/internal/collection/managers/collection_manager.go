package managers

import (
	collectionconst "backend/service-platform/app/internal/collection/constants/collection"
	"backend/service-platform/app/internal/collection/entities"
	collrepo "backend/service-platform/app/internal/collection/repositories"
	"backend/service-platform/app/internal/user/constants/currency"
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CollectionManager interface {
	ListCollections(ctx context.Context, filter ListCollectionsFilter) ([]CollectionWithItems, error)
	DeleteCollection(ctx context.Context, id uuid.UUID) error
}

type CollectionWithItems struct {
	Collection entity.Collection
	Items      []entity.CollectionItem
}

type ListCollectionsFilter struct {
	IDs              []uuid.UUID
	Names            []string
	Types            []collectionconst.Type
	RewardCurrencies []currency.Currency
	RewardAmounts    []int64
	MinRewardAmount  *int64
	MaxRewardAmount  *int64
	IsEnabled        *bool
	Limit            int
	Offset           int
}

type TransactionRunner interface {
	RunInTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx bun.Tx) error) error
}

type DefaultCollectionManager struct {
	collections collrepo.CollectionRepository
	items       collrepo.CollectionItemRepository
	txRunner    TransactionRunner
}

func NewCollectionManager(txRunner TransactionRunner, collections collrepo.CollectionRepository, items collrepo.CollectionItemRepository) CollectionManager {
	return &DefaultCollectionManager{
		collections: collections,
		items:       items,
		txRunner:    txRunner,
	}
}

func (m *DefaultCollectionManager) ListCollections(ctx context.Context, filter ListCollectionsFilter) ([]CollectionWithItems, error) {
	repoFilter := collrepo.CollectionFilter{
		IDs:              filter.IDs,
		Names:            filter.Names,
		Types:            filter.Types,
		RewardCurrencies: filter.RewardCurrencies,
		RewardAmounts:    filter.RewardAmounts,
		MinRewardAmount:  filter.MinRewardAmount,
		MaxRewardAmount:  filter.MaxRewardAmount,
		IsEnabled:        filter.IsEnabled,
		Limit:            filter.Limit,
		Offset:           filter.Offset,
	}

	collections, err := m.collections.List(ctx, repoFilter)
	if err != nil {
		return nil, err
	}
	if len(collections) == 0 {
		return []CollectionWithItems{}, nil
	}

	collectionIDs := make([]uuid.UUID, len(collections))
	for i, col := range collections {
		collectionIDs[i] = col.ID
	}
	items, err := m.items.ListByCollectionIDs(ctx, collectionIDs, false)
	if err != nil {
		return nil, err
	}
	grouped := make(map[uuid.UUID][]entity.CollectionItem, len(collections))
	for _, item := range items {
		grouped[item.CollectionID] = append(grouped[item.CollectionID], item)
	}

	result := make([]CollectionWithItems, len(collections))
	for i, col := range collections {
		result[i] = CollectionWithItems{
			Collection: col,
			Items:      grouped[col.ID],
		}
	}
	return result, nil
}

func (m *DefaultCollectionManager) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	err := m.txRunner.RunInTx(ctx, nil, func(txCtx context.Context, tx bun.Tx) error {
		if _, err := m.items.SoftDeleteByCollectionID(txCtx, tx, id); err != nil {
			return err
		}
		rows, err := m.collections.SoftDelete(txCtx, tx, id)
		if err != nil {
			return err
		}
		if rows == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
