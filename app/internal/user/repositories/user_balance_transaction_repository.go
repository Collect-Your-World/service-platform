package repository

import (
	"backend/service-platform/app/internal/platform/runtime"
	"backend/service-platform/app/internal/user/constants/currency"
	txconst "backend/service-platform/app/internal/user/constants/transaction"
	"backend/service-platform/app/internal/user/entities"
	"context"

	"github.com/google/uuid"
)

type UserBalanceTransactionRepository interface {
	Create(ctx context.Context, tx *entity.UserBalanceTransaction) (*entity.UserBalanceTransaction, error)
	ListByUser(ctx context.Context, userID uuid.UUID, _currency *currency.Currency, _type *txconst.TransactionType) ([]entity.UserBalanceTransaction, error)
}

type DefaultUserBalanceTransactionRepository struct {
	res runtime.Resource
}

func NewUserBalanceTransactionRepository(res runtime.Resource) UserBalanceTransactionRepository {
	return &DefaultUserBalanceTransactionRepository{res: res}
}

func (r *DefaultUserBalanceTransactionRepository) Create(ctx context.Context, t *entity.UserBalanceTransaction) (*entity.UserBalanceTransaction, error) {
	err := r.res.DB.NewInsert().Model(t).Returning("*").Scan(ctx, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *DefaultUserBalanceTransactionRepository) ListByUser(ctx context.Context, userID uuid.UUID, _currency *currency.Currency, _type *txconst.TransactionType) ([]entity.UserBalanceTransaction, error) {
	var items []entity.UserBalanceTransaction
	q := r.res.DB.ReplicaNewSelect().Model(&items).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		OrderExpr("created_at DESC")
	if _currency != nil {
		q = q.Where("currency = ?", *_currency)
	}
	if _type != nil {
		q = q.Where("type = ?", *_type)
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return items, nil
}
