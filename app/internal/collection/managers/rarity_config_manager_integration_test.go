package managers_test

import (
	integrationtest "backend/service-platform/app/internal/integrationtest"
	"context"
	"testing"
	"time"

	"backend/service-platform/app/internal/collection/entities"
	collrepo "backend/service-platform/app/internal/collection/repositories"
	"github.com/stretchr/testify/suite"
)

type RarityConfigManagerSuite struct {
	integrationtest.RouterSuite
}

func TestRarityConfigManagerSuite(t *testing.T) {
	suite.Run(t, new(RarityConfigManagerSuite))
}

func (s *RarityConfigManagerSuite) Test_CreateFindUpdateList_RarityConfigManager() {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	color := "#CAFEBABE"
	rc := &entity.RarityConfig{Code: "TEST", Label: "Test RC", Rank: 5, ColorHex: &color, DropWeight: 20}

	created, err := s.Managers.RarityConfigManager.Create(ctx, rc)
	s.R.NoError(err)
	s.A.NotZero(created.ID)

	// find by code
	found, err := s.Managers.RarityConfigManager.GetByCode(ctx, created.Code)
	s.R.NoError(err)
	s.A.Equal(created.ID, found.ID)

	// update
	found.Label = "Updated"
	updated, err := s.Managers.RarityConfigManager.Update(ctx, found)
	s.R.NoError(err)
	s.A.Equal("Updated", updated.Label)

	// list
	list, err := s.Managers.RarityConfigManager.List(ctx, collrepo.RarityConfigFilter{Codes: []string{created.Code}})
	s.R.NoError(err)
	s.A.Len(list, 1)
}
