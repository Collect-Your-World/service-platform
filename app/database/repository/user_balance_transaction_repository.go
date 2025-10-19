package repository

import (
	"backend/service-platform/app/database/entity"
	"backend/service-platform/app/internal/runtime"
	"context"
)

type UserBalanceTransactionRepository interface {
	Create(ctx context.Context, tx *entity.UserBalanceTransaction) (*entity.UserBalanceTransaction, error)
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
