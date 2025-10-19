package entity

import (
	"time"

	"backend/service-platform/app/database/constant/currency"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserBalance struct {
	bun.BaseModel `bun:"table:user_balances,alias:ub"`

	ID        uuid.UUID         `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID    uuid.UUID         `bun:"user_id,notnull,type:uuid"`
	Balance   int64             `bun:"balance,notnull"`
	Currency  currency.Currency `bun:"currency,notnull"`
	CreatedAt time.Time         `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt *time.Time        `bun:"updated_at"`
	DeletedAt *time.Time        `bun:"deleted_at,soft_delete"`
}

func (u UserBalance) Alias() string { return "ub" }
