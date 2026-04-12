package managers_test

import (
	collectionconst "backend/service-platform/app/internal/collection/constants/collection"
	itemconst "backend/service-platform/app/internal/collection/constants/item"
	"backend/service-platform/app/internal/collection/entities"
	integrationtest "backend/service-platform/app/internal/integrationtest"
	"backend/service-platform/app/internal/user/constants/currency"
	"context"
	"testing"
	"time"

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

	color := "#123456"
	rc := entity.RarityConfig{Code: "M" + uuid.New().String()[:8], Label: "Common", Rank: 1, ColorHex: &color, DropWeight: 10}
	createdRC, err := s.Repositories.RarityConfigRepository.Create(ctx, &rc)
	s.R.NoError(err)

	it := &entity.Item{
		Name:           "Mgr Item",
		Slug:           "mgr-item-" + uuid.New().String(),
		Description:    nil,
		ItemType:       itemconst.Landscape,
		RarityConfigID: createdRC.ID,
	}
	created, err := s.Managers.ItemManager.CreateItem(ctx, it)
	s.R.NoError(err)
	s.A.Equal("Mgr Item", created.Name)

	got, err := s.Managers.ItemManager.GetItem(ctx, created.ID)
	s.R.NoError(err)
	s.A.Equal(created.ID, got.ID)

	got.Name = "Mgr Item Updated"
	updated, err := s.Managers.ItemManager.UpdateItem(ctx, got)
	s.R.NoError(err)
	s.A.Equal("Mgr Item Updated", updated.Name)

	err = s.Managers.ItemManager.DeleteItems(ctx, []uuid.UUID{created.ID})
	s.R.NoError(err)

	var soft entity.Item
	err = s.Resource.DB.ReplicaNewSelect().Model(&soft).Where("id = ?", created.ID).WhereAllWithDeleted().Scan(ctx)
	s.R.NoError(err)
	s.A.NotNil(soft.DeletedAt)
}

func (s *ItemManagerSuite) Test_ItemManager_WithCollectionViaJoin() {
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	col := entity.Collection{
		Name:           "ItemMgr Collection",
		Slug:           "item-mgr-col-" + uuid.New().String(),
		Type:           collectionconst.Global,
		RewardAmount:   0,
		RewardCurrency: currency.COIN,
		IsEnabled:      true,
	}
	createdCol, err := s.Repositories.CollectionRepository.Create(ctx, &col)
	s.R.NoError(err)

	color := "#123456"
	rc := entity.RarityConfig{Code: "J" + uuid.New().String()[:8], Label: "Common", Rank: 1, ColorHex: &color, DropWeight: 10}
	createdRC, err := s.Repositories.RarityConfigRepository.Create(ctx, &rc)
	s.R.NoError(err)

	it := &entity.Item{
		Name:           "Joined Item",
		Slug:           "joined-" + uuid.New().String(),
		ItemType:       itemconst.Food,
		RarityConfigID: createdRC.ID,
	}
	createdItem, err := s.Managers.ItemManager.CreateItem(ctx, it)
	s.R.NoError(err)

	_, err = s.Repositories.CollectionItemRepository.Create(ctx, &entity.CollectionItem{
		CollectionID: createdCol.ID,
		ItemID:       createdItem.ID,
	})
	s.R.NoError(err)

	listed, err := s.Repositories.CollectionItemRepository.ListByCollectionIDs(ctx, []uuid.UUID{createdCol.ID}, false)
	s.R.NoError(err)
	s.A.Len(listed, 1)
	s.A.Equal(createdItem.ID, listed[0].ItemID)
}
