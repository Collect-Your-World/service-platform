package entity

import (
	"time"

	"backend/service-platform/app/database/constant/currency"
	txconst "backend/service-platform/app/database/constant/transaction"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserBalanceTransaction struct {
	bun.BaseModel `bun:"table:user_balance_transactions,alias:ubt"`

	ID        uuid.UUID               `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID    uuid.UUID               `bun:"user_id,notnull,type:uuid"`
	Amount    int64                   `bun:"amount,notnull"`
	Currency  currency.Currency       `bun:"currency,notnull"`
	Type      txconst.TransactionType `bun:"type,notnull"`
	Source    txconst.Source          `bun:"source,notnull"`
	Status    txconst.Status          `bun:"status,notnull"`
	CreatedAt time.Time               `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt *time.Time              `bun:"updated_at"`
	DeletedAt *time.Time              `bun:"deleted_at,soft_delete"`
}

func (u UserBalanceTransaction) Alias() string { return "ubt" }
