package managers

import (
	"backend/service-platform/app/internal/user/constants/currency"
	txconst "backend/service-platform/app/internal/user/constants/transaction"
	"backend/service-platform/app/internal/user/entities"
	userrepo "backend/service-platform/app/internal/user/repositories"
	"context"

	"github.com/google/uuid"
)

type UserBalanceManager interface {
	RecordChange(ctx context.Context, userID uuid.UUID, cur currency.Currency, delta int64, txType txconst.TransactionType, source txconst.Source, status txconst.Status) (*entity.UserBalance, *entity.UserBalanceTransaction, error)
	GetAllBalances(ctx context.Context, userID uuid.UUID) (map[string]int64, error)
	GetBalanceByCurrency(ctx context.Context, userID uuid.UUID, cur currency.Currency) (map[string]int64, error)
	GetHistory(ctx context.Context, userID uuid.UUID, _currency *currency.Currency, _type *txconst.TransactionType) ([]entity.UserBalanceTransaction, error)
}

type DefaultUserBalanceManager struct {
	balancesRepo     userrepo.UserBalanceRepository
	transactionsRepo userrepo.UserBalanceTransactionRepository
}

func NewUserBalanceManager(balances userrepo.UserBalanceRepository, txs userrepo.UserBalanceTransactionRepository) UserBalanceManager {
	return &DefaultUserBalanceManager{balancesRepo: balances, transactionsRepo: txs}
}

func (m *DefaultUserBalanceManager) RecordChange(ctx context.Context, userID uuid.UUID, cur currency.Currency, delta int64, txType txconst.TransactionType, source txconst.Source, status txconst.Status) (*entity.UserBalance, *entity.UserBalanceTransaction, error) {
	updatedBalance, err := m.balancesRepo.UpsertAndAddDelta(ctx, userID, cur, delta)
	if err != nil {
		return nil, nil, err
	}

	tx := &entity.UserBalanceTransaction{
		UserID:   userID,
		Amount:   delta,
		Currency: cur,
		Type:     txType,
		Source:   source,
		Status:   status,
	}
	createdTx, err := m.transactionsRepo.Create(ctx, tx)
	if err != nil {
		return nil, nil, err
	}
	return updatedBalance, createdTx, nil
}

func (m *DefaultUserBalanceManager) GetAllBalances(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {
	items, err := m.balancesRepo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make(map[string]int64, len(items))
	for _, it := range items {
		res[string(it.Currency)] = it.Balance
	}
	return res, nil
}

func (m *DefaultUserBalanceManager) GetBalanceByCurrency(ctx context.Context, userID uuid.UUID, cur currency.Currency) (map[string]int64, error) {
	ub, err := m.balancesRepo.FindByUserAndCurrency(ctx, userID, cur)
	if err != nil {
		return map[string]int64{string(cur): 0}, nil
	}
	return map[string]int64{string(cur): ub.Balance}, nil
}

func (m *DefaultUserBalanceManager) GetHistory(ctx context.Context, userID uuid.UUID, _currency *currency.Currency, _type *txconst.TransactionType) ([]entity.UserBalanceTransaction, error) {
	return m.transactionsRepo.ListByUser(ctx, userID, _currency, _type)
}
