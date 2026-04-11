package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CollectionItem struct {
	bun.BaseModel `bun:"table:collection_items,alias:ci"`

	ID           uuid.UUID  `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	CollectionID uuid.UUID  `bun:"collection_id,notnull,type:uuid"`
	ItemID       uuid.UUID  `bun:"item_id,notnull,type:uuid"`
	ImageURL     *string    `bun:"image_url"`
	CreatedAt    time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt    *time.Time `bun:"updated_at"`
	DeletedAt    *time.Time `bun:"deleted_at,soft_delete"`
}

func (CollectionItem) Alias() string {
	return "ci"
}
