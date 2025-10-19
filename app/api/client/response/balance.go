package response

import (
	"backend/service-platform/app/database/constant/currency"
	txconst "backend/service-platform/app/database/constant/transaction"
	"time"

	"github.com/google/uuid"
)

type BalanceTransactionResponse struct {
	ID        uuid.UUID         `json:"id"`
	Amount    int64             `json:"amount"`
	Currency  currency.Currency `json:"currency"`
	Type      txconst.Type      `json:"type"`
	Source    txconst.Source    `json:"source"`
	Status    txconst.Status    `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
}
