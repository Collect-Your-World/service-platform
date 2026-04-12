package models

import (
	"time"

	"github.com/google/uuid"
)

type BalanceTransactionResponse struct {
	ID            uuid.UUID              `json:"id"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	Type          string                 `json:"type"`
	Source        string                 `json:"source"`
	Status        string                 `json:"status"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	BalanceBefore int64                  `json:"balance_before"`
	BalanceAfter  int64                  `json:"balance_after"`
	CreatedAt     time.Time              `json:"created_at"`
}
