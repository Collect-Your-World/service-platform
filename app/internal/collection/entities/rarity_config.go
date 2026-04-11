package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type RarityConfig struct {
	bun.BaseModel `bun:"table:rarity_configs,alias:rc"`

	ID         uuid.UUID  `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Code       string     `bun:"code,notnull,unique"`
	Label      string     `bun:"label,notnull"`
	Rank       int16      `bun:"rank,notnull"`
	ColorHex   *string    `bun:"color_hex"`
	DropWeight int        `bun:"drop_weight,notnull,default:0"`
	CreatedAt  time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt  *time.Time `bun:"updated_at"`
	DeletedAt  *time.Time `bun:"deleted_at,soft_delete"`
}

func (RarityConfig) Alias() string {
	return "rc"
}
