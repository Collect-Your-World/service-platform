package response

import (
	"time"

	"github.com/google/uuid"
)

type BalanceTransactionResponse struct {
	ID        uuid.UUID `json:"id"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
