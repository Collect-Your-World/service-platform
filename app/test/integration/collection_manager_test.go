package integration

import (
	collectionconst "backend/service-platform/app/internal/collection/constants/collection"
	"backend/service-platform/app/internal/collection/entities"
	collmanagers "backend/service-platform/app/internal/collection/managers"
	collrepo "backend/service-platform/app/internal/collection/repositories"
	"backend/service-platform/app/internal/user/constants/currency"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type CollectionManagerSuite struct {
	RouterSuite
}

func TestCollectionManagerSuite(t *testing.T) {
	suite.Run(t, new(CollectionManagerSuite))
}

func (s *CollectionManagerSuite) Test_ListCollections_WithFilters() {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	colA := s.seedCollection(ctx, entity.Collection{
		Name:           "Northern Adventures",
		Type:           collectionconst.Theme,
		RewardAmount:   100,
		RewardCurrency: currency.COIN,
		IsEnabled:      true,
	}, []entity.CollectionItem{
		{ItemID: uuid.New()},
		{ItemID: uuid.New()},
	})

	colB := s.seedCollection(ctx, entity.Collection{
		Name:           "Pacific Trails",
		Type:           collectionconst.Location,
		RewardAmount:   500,
		RewardCurrency: currency.SPIN,
		IsEnabled:      false,
	}, []entity.CollectionItem{
		{ItemID: uuid.New()},
	})

	result, err := s.managers.CollectionManager.ListCollections(ctx, collmanagers.ListCollectionsFilter{
		Types:            []collectionconst.Type{collectionconst.Theme},
		RewardCurrencies: []currency.Currency{currency.COIN},
		IsEnabled:        boolPtr(true),
	})
	s.r.NoError(err)
	s.a.Len(result, 1)
	s.a.Equal(colA.ID, result[0].Collection.ID)
	s.a.Len(result[0].Items, 2)

	byName, err := s.managers.CollectionManager.ListCollections(ctx, collmanagers.ListCollectionsFilter{
		Names: []string{colB.Name},
	})
	s.r.NoError(err)
	s.a.Len(byName, 1)
	s.a.Equal(colB.ID, byName[0].Collection.ID)
	s.a.Len(byName[0].Items, 1)
}

func (s *CollectionManagerSuite) Test_DeleteCollection_SoftDeletesCascade() {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	col := s.seedCollection(ctx, entity.Collection{
		Name:           "Global Treasures",
		Type:           collectionconst.Global,
		RewardAmount:   750,
		RewardCurrency: currency.COIN,
		IsEnabled:      true,
	}, []entity.CollectionItem{
		{ItemID: uuid.New()},
		{ItemID: uuid.New()},
	})

	err := s.managers.CollectionManager.DeleteCollection(ctx, col.ID)
	s.r.NoError(err)

	collections, err := s.repositories.CollectionRepository.List(ctx, collrepo.CollectionFilter{
		IDs:            []uuid.UUID{col.ID},
		IncludeDeleted: true,
	})
	s.r.NoError(err)
	s.a.Len(collections, 1)
	s.a.NotNil(collections[0].DeletedAt)

	items, err := s.repositories.CollectionItemRepository.ListByCollectionIDs(ctx, []uuid.UUID{col.ID}, true)
	s.r.NoError(err)
	s.a.Len(items, 2)
	for _, it := range items {
		s.a.NotNil(it.DeletedAt)
	}
}

func (s *CollectionManagerSuite) Test_DeleteCollection_ReturnsErrorWhenMissing() {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	err := s.managers.CollectionManager.DeleteCollection(ctx, uuid.New())
	s.r.Error(err)
}

func (s *CollectionManagerSuite) seedCollection(ctx context.Context, collection entity.Collection, items []entity.CollectionItem) *entity.Collection {
	created, err := s.repositories.CollectionRepository.Create(ctx, &collection)
	s.r.NoError(err)

	for i := range items {
		items[i].CollectionID = created.ID
		_, err := s.repositories.CollectionItemRepository.Create(ctx, &items[i])
		s.r.NoError(err)
	}

	return created
}

func boolPtr(value bool) *bool {
	return &value
}
