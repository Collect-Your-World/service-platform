package entity

import (
	collectionconst "backend/service-platform/app/internal/collection/constants/collection"
	"backend/service-platform/app/internal/user/constants/currency"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Collection struct {
	bun.BaseModel `bun:"table:collections,alias:c"`

	ID             uuid.UUID            `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name           string               `bun:"name,notnull"`
	Slug           string               `bun:"slug,notnull"`
	Description    *string              `bun:"description"`
	Type           collectionconst.Type `bun:"collection_type,notnull"`
	RewardAmount   int64                `bun:"reward_amount,notnull,default:0"`
	RewardCurrency currency.Currency    `bun:"reward_currency,notnull"`
	IsEnabled      bool                 `bun:"is_enabled,notnull,default:true"`
	CreatedAt      time.Time            `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      *time.Time           `bun:"updated_at"`
	DeletedAt      *time.Time           `bun:"deleted_at,soft_delete"`
	Items          []CollectionItem     `bun:"rel:has-many,join:id=collection_id"`
}

func (Collection) Alias() string {
	return "c"
}
