package managers_test

import (
	collectionconst "backend/service-platform/app/internal/collection/constants/collection"
	itemconst "backend/service-platform/app/internal/collection/constants/item"
	entity "backend/service-platform/app/internal/collection/entities"
	integrationtest "backend/service-platform/app/internal/integrationtest"
	"backend/service-platform/app/internal/user/constants/currency"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ItemsIntegrationSuite struct {
	integrationtest.RouterSuite
}

func TestItemsIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ItemsIntegrationSuite))
}

func (s *ItemsIntegrationSuite) Test_CreateAndRetrieveItem_WithRarityConfig() {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	color := "#FFFFFF"
	rc := entity.RarityConfig{
		Code:       "C" + uuid.New().String()[:8],
		Label:      "Test Common",
		Rank:       1,
		ColorHex:   &color,
		DropWeight: 100,
	}
	createdRC, err := s.Repositories.RarityConfigRepository.Create(ctx, &rc)
	s.R.NoError(err)

	item := entity.Item{
		Name:           "Integration Item",
		Slug:           "integration-item-" + uuid.New().String(),
		Description:    nil,
		ItemType:       itemconst.Other,
		RarityConfigID: createdRC.ID,
		ImageURL:       nil,
		CountryID:      nil,
		LocationID:     nil,
	}

	createdItem, err := s.Repositories.ItemRepository.Create(ctx, &item)
	s.R.NoError(err)
	s.A.Equal("Integration Item", createdItem.Name)

	fetched, err := s.Repositories.ItemRepository.FindByID(ctx, createdItem.ID)
	s.R.NoError(err)
	s.A.Equal(createdItem.ID, fetched.ID)
	s.A.Equal(createdItem.Slug, fetched.Slug)
}

func (s *ItemsIntegrationSuite) Test_CollectionRepository_WithSlug() {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	slug := "theme-" + uuid.New().String()
	col := entity.Collection{
		Name:           "Themed",
		Slug:           slug,
		Type:           collectionconst.Theme,
		RewardAmount:   1,
		RewardCurrency: currency.COIN,
		IsEnabled:      true,
	}
	created, err := s.Repositories.CollectionRepository.Create(ctx, &col)
	s.R.NoError(err)

	found, err := s.Repositories.CollectionRepository.FindBySlug(ctx, slug)
	s.R.NoError(err)
	s.A.Equal(created.ID, found.ID)
}
