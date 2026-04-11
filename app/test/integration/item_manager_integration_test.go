package integration

import (
	"context"
	"testing"
	"time"

	collectionconst "backend/service-platform/app/database/constant/collection"
	"backend/service-platform/app/database/constant/currency"
	"backend/service-platform/app/database/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ItemManagerSuite struct {
	RouterSuite
}

func TestItemManagerSuite(t *testing.T) {
	suite.Run(t, new(ItemManagerSuite))
}

func (s *ItemManagerSuite) Test_CreateGetUpdateDelete_ItemManager() {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()
	// check if test DB has collection_id column; if not, skip manager-level test
	var tmp int
	q := "SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='items' AND column_name='collection_id'"
	err := s.resource.DB.PrimaryDb.QueryRowContext(ctx, q).Scan(&tmp)
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
	createdCol, err := s.repositories.CollectionRepository.Create(ctx, &col)
	s.r.NoError(err)

	// create rarity config
	color := "#123456"
	rc := entity.RarityConfig{Code: "COMMON", Label: "Common", Rank: 1, ColorHex: &color, DropWeight: 10}
	createdRC, err := s.repositories.RarityConfigRepository.Create(ctx, &rc)
	s.r.NoError(err)

	// create item via manager
	it := &entity.Item{
		Name:           "Mgr Item",
		Description:    nil,
		RarityConfigID: &createdRC.ID,
		CollectionID:   &createdCol.ID,
	}
	created, err := s.managers.ItemManager.CreateItem(ctx, it)
	s.r.NoError(err)
	s.a.Equal("Mgr Item", created.Name)

	// get
	got, err := s.managers.ItemManager.GetItem(ctx, created.ID)
	s.r.NoError(err)
	s.a.Equal(created.ID, got.ID)

	// update
	got.Name = "Mgr Item Updated"
	updated, err := s.managers.ItemManager.UpdateItem(ctx, got)
	s.r.NoError(err)
	s.a.Equal("Mgr Item Updated", updated.Name)

	// list by collection
	listed, err := s.managers.ItemManager.ListItemsByCollectionIDs(ctx, []uuid.UUID{createdCol.ID}, false)
	s.r.NoError(err)
	s.a.True(len(listed) >= 1)

	// delete
	err = s.managers.ItemManager.DeleteItems(ctx, []uuid.UUID{created.ID})
	s.r.NoError(err)

	// ensure deleted (soft)
	listedDeleted, err := s.repositories.ItemRepository.ListByCollectionIDs(ctx, []uuid.UUID{createdCol.ID}, true)
	s.r.NoError(err)
	found := false
	for _, it := range listedDeleted {
		if it.ID == created.ID {
			found = true
			s.a.NotNil(it.DeletedAt)
		}
	}
	s.a.True(found)
}
