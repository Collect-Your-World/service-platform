package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"backend/service-platform/app/internal/user/constants/currency"
	txconst "backend/service-platform/app/internal/user/constants/transaction"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type TransactionMetadata map[string]interface{}

func (m TransactionMetadata) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TransactionMetadata: %w", err)
	}
	return string(data), nil
}

func (m *TransactionMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into TransactionMetadata", value)
	}
	if len(bytes) == 0 {
		*m = TransactionMetadata{}
		return nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return err
	}
	*m = raw
	return nil
}

type UserBalanceTransaction struct {
	bun.BaseModel `bun:"table:user_balance_transactions,alias:ubt"`

	ID            uuid.UUID               `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID        uuid.UUID               `bun:"user_id,notnull,type:uuid"`
	Amount        int64                   `bun:"amount,notnull"`
	Currency      currency.Currency       `bun:"currency,notnull"`
	Type          txconst.TransactionType `bun:"type,notnull"`
	Source        txconst.Source          `bun:"source,notnull"`
	Status        txconst.Status          `bun:"status,notnull"`
	Metadata      TransactionMetadata     `bun:"metadata,type:jsonb"`
	BalanceBefore int64                   `bun:"balance_before,notnull"`
	BalanceAfter  int64                   `bun:"balance_after,notnull"`
	CreatedAt     time.Time               `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     *time.Time              `bun:"updated_at"`
	DeletedAt     *time.Time              `bun:"deleted_at,soft_delete"`
}

func (u UserBalanceTransaction) Alias() string { return "ubt" }
