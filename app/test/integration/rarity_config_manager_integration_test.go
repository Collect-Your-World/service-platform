package integration

import (
	"context"
	"testing"
	"time"

	"backend/service-platform/app/internal/collection/entities"
	collrepo "backend/service-platform/app/internal/collection/repositories"
	"github.com/stretchr/testify/suite"
)

type RarityConfigManagerSuite struct {
	RouterSuite
}

func TestRarityConfigManagerSuite(t *testing.T) {
	suite.Run(t, new(RarityConfigManagerSuite))
}

func (s *RarityConfigManagerSuite) Test_CreateFindUpdateList_RarityConfigManager() {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	color := "#CAFEBABE"
	rc := &entity.RarityConfig{Code: "TEST", Label: "Test RC", Rank: 5, ColorHex: &color, DropWeight: 20}

	created, err := s.managers.RarityConfigManager.Create(ctx, rc)
	s.r.NoError(err)
	s.a.NotZero(created.ID)

	// find by code
	found, err := s.managers.RarityConfigManager.GetByCode(ctx, created.Code)
	s.r.NoError(err)
	s.a.Equal(created.ID, found.ID)

	// update
	found.Label = "Updated"
	updated, err := s.managers.RarityConfigManager.Update(ctx, found)
	s.r.NoError(err)
	s.a.Equal("Updated", updated.Label)

	// list
	list, err := s.managers.RarityConfigManager.List(ctx, collrepo.RarityConfigFilter{Codes: []string{created.Code}})
	s.r.NoError(err)
	s.a.Len(list, 1)
}
