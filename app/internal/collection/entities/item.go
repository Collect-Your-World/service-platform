package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Item struct {
	bun.BaseModel `bun:"table:items,alias:i"`

	ID             uuid.UUID  `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name           string     `bun:"name,notnull"`
	Description    *string    `bun:"description"`
	RarityConfigID *uuid.UUID `bun:"rarity_config_id,type:uuid"`
	ImageURL       *string    `bun:"image_url"`
	CountryID      *uuid.UUID `bun:"country_id,type:uuid"`
	LocationID     *uuid.UUID `bun:"location_id,type:uuid"`
	CollectionID   *uuid.UUID `bun:"collection_id,type:uuid"`
	CreatedAt      time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      *time.Time `bun:"updated_at"`
	DeletedAt      *time.Time `bun:"deleted_at,soft_delete"`
}

func (Item) Alias() string {
	return "i"
}
