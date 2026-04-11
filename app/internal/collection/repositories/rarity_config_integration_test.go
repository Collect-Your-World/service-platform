package repository_test

import (
	integrationtest "backend/service-platform/app/internal/integrationtest"
	"context"
	"testing"
	"time"

	entity "backend/service-platform/app/internal/collection/entities"
	collrepo "backend/service-platform/app/internal/collection/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type RarityConfigIntegrationSuite struct {
	integrationtest.RouterSuite
}

func TestRarityConfigIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RarityConfigIntegrationSuite))
}

func (s *RarityConfigIntegrationSuite) Test_CreateFindUpdateList_RarityConfig() {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	color := "#ABCDEF"
	rc := entity.RarityConfig{
		Code:       "COMMON",
		Label:      "Common",
		Rank:       1,
		ColorHex:   &color,
		DropWeight: 50,
	}

	// Create
	created, err := s.Repositories.RarityConfigRepository.Create(ctx, &rc)
	s.R.NoError(err)
	s.A.NotZero(created.ID)

	// Find by code
	found, err := s.Repositories.RarityConfigRepository.FindByCode(ctx, created.Code)
	s.R.NoError(err)
	s.A.Equal(created.ID, found.ID)

	// Update label
	found.Label = "Common Updated"
	updated, err := s.Repositories.RarityConfigRepository.Update(ctx, found)
	s.R.NoError(err)
	s.A.Equal("Common Updated", updated.Label)

	// List
	list, err := s.Repositories.RarityConfigRepository.List(ctx, collrepo.RarityConfigFilter{Codes: []string{created.Code}})
	s.R.NoError(err)
	s.A.Len(list, 1)
	s.A.Equal(created.ID, list[0].ID)

	// sanity: GetByID
	got, err := s.Repositories.RarityConfigRepository.FindByID(ctx, created.ID)
	s.R.NoError(err)
	s.A.Equal(created.ID, got.ID)

	// ensure cleanup by deleting created directly (soft delete)
	// Not strictly necessary — RouterSuite teardown will clean created rows by timestamp
	_ = created
	_ = uuid.New()
}
