package integration

import (
	"context"
	"testing"
	"time"

	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/database/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type RarityConfigIntegrationSuite struct {
	RouterSuite
}

func TestRarityConfigIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RarityConfigIntegrationSuite))
}

func (s *RarityConfigIntegrationSuite) Test_CreateFindUpdateList_RarityConfig() {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
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
	created, err := s.repositories.RarityConfigRepository.Create(ctx, &rc)
	s.r.NoError(err)
	s.a.NotZero(created.ID)

	// Find by code
	found, err := s.repositories.RarityConfigRepository.FindByCode(ctx, created.Code)
	s.r.NoError(err)
	s.a.Equal(created.ID, found.ID)

	// Update label
	found.Label = "Common Updated"
	updated, err := s.repositories.RarityConfigRepository.Update(ctx, found)
	s.r.NoError(err)
	s.a.Equal("Common Updated", updated.Label)

	// List
	list, err := s.repositories.RarityConfigRepository.List(ctx, repository.RarityConfigFilter{Codes: []string{created.Code}})
	s.r.NoError(err)
	s.a.Len(list, 1)
	s.a.Equal(created.ID, list[0].ID)

	// sanity: GetByID
	got, err := s.repositories.RarityConfigRepository.FindByID(ctx, created.ID)
	s.r.NoError(err)
	s.a.Equal(created.ID, got.ID)

	// ensure cleanup by deleting created directly (soft delete)
	// Not strictly necessary — RouterSuite teardown will clean created rows by timestamp
	_ = created
	_ = uuid.New()
}
