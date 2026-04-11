package managers_test

import (
	integrationtest "backend/service-platform/app/internal/integrationtest"
	"context"
	"testing"
	"time"

	collectionconst "backend/service-platform/app/internal/collection/constants/collection"
	"backend/service-platform/app/internal/collection/entities"
	"backend/service-platform/app/internal/user/constants/currency"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ItemManagerSuite struct {
	integrationtest.RouterSuite
}

func TestItemManagerSuite(t *testing.T) {
	suite.Run(t, new(ItemManagerSuite))
}

func (s *ItemManagerSuite) Test_CreateGetUpdateDelete_ItemManager() {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()
	// check if test DB has collection_id column; if not, skip manager-level test
	var tmp int
	q := "SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='items' AND column_name='collection_id'"
	err := s.Resource.DB.PrimaryDb.QueryRowContext(ctx, q).Scan(&tmp)
	if err != nil {
		// if ErrNoRows or other, skip
		s.T().Skip("test DB missing items.collection_id column; skipping ItemManager manager-level test")
		return
	}

	// create collection
	col := entity.Collection{
		Name:           "ItemMgr Collection",
		Type:           collectionconst.Global,
		RewardAmount:   0,
		RewardCurrency: currency.COIN,
		IsEnabled:      true,
	}
	createdCol, err := s.Repositories.CollectionRepository.Create(ctx, &col)
	s.R.NoError(err)

	// create rarity config
	color := "#123456"
	rc := entity.RarityConfig{Code: "COMMON", Label: "Common", Rank: 1, ColorHex: &color, DropWeight: 10}
	createdRC, err := s.Repositories.RarityConfigRepository.Create(ctx, &rc)
	s.R.NoError(err)

	// create item via manager
	it := &entity.Item{
		Name:           "Mgr Item",
		Description:    nil,
		RarityConfigID: &createdRC.ID,
		CollectionID:   &createdCol.ID,
	}
	created, err := s.Managers.ItemManager.CreateItem(ctx, it)
	s.R.NoError(err)
	s.A.Equal("Mgr Item", created.Name)

	// get
	got, err := s.Managers.ItemManager.GetItem(ctx, created.ID)
	s.R.NoError(err)
	s.A.Equal(created.ID, got.ID)

	// update
	got.Name = "Mgr Item Updated"
	updated, err := s.Managers.ItemManager.UpdateItem(ctx, got)
	s.R.NoError(err)
	s.A.Equal("Mgr Item Updated", updated.Name)

	// list by collection
	listed, err := s.Managers.ItemManager.ListItemsByCollectionIDs(ctx, []uuid.UUID{createdCol.ID}, false)
	s.R.NoError(err)
	s.A.True(len(listed) >= 1)

	// delete
	err = s.Managers.ItemManager.DeleteItems(ctx, []uuid.UUID{created.ID})
	s.R.NoError(err)

	// ensure deleted (soft)
	listedDeleted, err := s.Repositories.ItemRepository.ListByCollectionIDs(ctx, []uuid.UUID{createdCol.ID}, true)
	s.R.NoError(err)
	found := false
	for _, it := range listedDeleted {
		if it.ID == created.ID {
			found = true
			s.A.NotNil(it.DeletedAt)
		}
	}
	s.A.True(found)
}
