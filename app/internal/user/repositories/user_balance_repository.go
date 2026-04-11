package repository

import (
	"backend/service-platform/app/internal/platform/runtime"
	"backend/service-platform/app/internal/user/constants/currency"
	"backend/service-platform/app/internal/user/entities"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserBalanceRepository interface {
	Create(ctx context.Context, balance *entity.UserBalance) (*entity.UserBalance, error)
	FindByUserAndCurrency(ctx context.Context, userID uuid.UUID, currency currency.Currency) (*entity.UserBalance, error)
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]entity.UserBalance, error)
	UpsertAndAddDelta(ctx context.Context, userID uuid.UUID, currency currency.Currency, delta int64) (*entity.UserBalance, error)
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

type DefaultUserBalanceRepository struct {
	res runtime.Resource
}

func NewUserBalanceRepository(res runtime.Resource) UserBalanceRepository {
	return &DefaultUserBalanceRepository{res: res}
}

func (r *DefaultUserBalanceRepository) Create(ctx context.Context, balance *entity.UserBalance) (*entity.UserBalance, error) {
	err := r.res.DB.NewInsert().Model(balance).Returning("*").Scan(ctx, balance)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (r *DefaultUserBalanceRepository) FindByUserAndCurrency(ctx context.Context, userID uuid.UUID, currency currency.Currency) (*entity.UserBalance, error) {
	ub := new(entity.UserBalance)
	err := r.res.DB.ReplicaNewSelect().Model(ub).
		Where("user_id = ?", userID).
		Where("currency = ?", currency).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return ub, nil
}

func (r *DefaultUserBalanceRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]entity.UserBalance, error) {
	var items []entity.UserBalance
	err := r.res.DB.ReplicaNewSelect().Model(&items).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DefaultUserBalanceRepository) UpsertAndAddDelta(ctx context.Context, userID uuid.UUID, currency currency.Currency, delta int64) (*entity.UserBalance, error) {
	var updated *entity.UserBalance
	err := r.res.DB.RunInTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable}, func(txCtx context.Context, tx bun.Tx) error {
		var current entity.UserBalance
		selErr := tx.NewSelect().Model(&current).
			Where("user_id = ?", userID).
			Where("currency = ?", currency).
			Where("deleted_at IS NULL").
			For("UPDATE").
			Scan(txCtx)
		if selErr != nil {
			// if no rows found, create a new one
			if errors.Is(selErr, sql.ErrNoRows) {
				newItem := &entity.UserBalance{UserID: userID, Currency: currency, Balance: delta}
				if err := tx.NewInsert().Model(newItem).Returning("*").Scan(txCtx, newItem); err != nil {
					return err
				}
				updated = newItem
				return nil
			}
			return selErr
		}

		current.Balance = current.Balance + delta
		var out entity.UserBalance
		if err := tx.NewUpdate().Model(&current).WherePK().Where("deleted_at IS NULL").Returning("*").Scan(txCtx, &out); err != nil {
			return err
		}
		updated = &out
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *DefaultUserBalanceRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	res, err := r.res.DB.NewUpdate().Model((*entity.UserBalance)(nil)).
		Set("deleted_at = NOW()").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}
