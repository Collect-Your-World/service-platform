package managers_test

import (
	collectionconst "backend/service-platform/app/internal/collection/constants/collection"
	entity "backend/service-platform/app/internal/collection/entities"
	collmanagers "backend/service-platform/app/internal/collection/managers"
	collrepo "backend/service-platform/app/internal/collection/repositories"
	integrationtest "backend/service-platform/app/internal/integrationtest"
	"backend/service-platform/app/internal/user/constants/currency"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type CollectionManagerSuite struct {
	integrationtest.RouterSuite
}

func TestCollectionManagerSuite(t *testing.T) {
	suite.Run(t, new(CollectionManagerSuite))
}

func (s *CollectionManagerSuite) Test_ListCollections_WithFilters() {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
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

	result, err := s.Managers.CollectionManager.ListCollections(ctx, collmanagers.ListCollectionsFilter{
		Types:            []collectionconst.Type{collectionconst.Theme},
		RewardCurrencies: []currency.Currency{currency.COIN},
		IsEnabled:        boolPtr(true),
	})
	s.R.NoError(err)
	s.A.Len(result, 1)
	s.A.Equal(colA.ID, result[0].Collection.ID)
	s.A.Len(result[0].Items, 2)

	byName, err := s.Managers.CollectionManager.ListCollections(ctx, collmanagers.ListCollectionsFilter{
		Names: []string{colB.Name},
	})
	s.R.NoError(err)
	s.A.Len(byName, 1)
	s.A.Equal(colB.ID, byName[0].Collection.ID)
	s.A.Len(byName[0].Items, 1)
}

func (s *CollectionManagerSuite) Test_DeleteCollection_SoftDeletesCascade() {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
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

	err := s.Managers.CollectionManager.DeleteCollection(ctx, col.ID)
	s.R.NoError(err)

	collections, err := s.Repositories.CollectionRepository.List(ctx, collrepo.CollectionFilter{
		IDs:            []uuid.UUID{col.ID},
		IncludeDeleted: true,
	})
	s.R.NoError(err)
	s.A.Len(collections, 1)
	s.A.NotNil(collections[0].DeletedAt)

	items, err := s.Repositories.CollectionItemRepository.ListByCollectionIDs(ctx, []uuid.UUID{col.ID}, true)
	s.R.NoError(err)
	s.A.Len(items, 2)
	for _, it := range items {
		s.A.NotNil(it.DeletedAt)
	}
}

func (s *CollectionManagerSuite) Test_DeleteCollection_ReturnsErrorWhenMissing() {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	err := s.Managers.CollectionManager.DeleteCollection(ctx, uuid.New())
	s.R.Error(err)
}

func (s *CollectionManagerSuite) seedCollection(ctx context.Context, collection entity.Collection, items []entity.CollectionItem) *entity.Collection {
	created, err := s.Repositories.CollectionRepository.Create(ctx, &collection)
	s.R.NoError(err)

	for i := range items {
		items[i].CollectionID = created.ID
		_, err := s.Repositories.CollectionItemRepository.Create(ctx, &items[i])
		s.R.NoError(err)
	}

	return created
}

func boolPtr(value bool) *bool {
	return &value
}
