package repository

import (
	"backend/service-platform/app/internal/runtime"
)

type Repositories struct {
	UserRepository                   UserRepository
	SessionRepository                SessionRepository
	JobRepository                    JobRepository
	UserBalanceRepository            UserBalanceRepository
	UserBalanceTransactionRepository UserBalanceTransactionRepository
	RewardRepository                 RewardRepository
	UserRewardRepository             UserRewardRepository
}

func NewRepositories(res runtime.Resource) *Repositories {
	return &Repositories{
		UserRepository:                   NewUserRepository(res),
		SessionRepository:                NewSessionRepository(res),
		JobRepository:                    NewJobRepository(res),
		UserBalanceRepository:            NewUserBalanceRepository(res),
		UserBalanceTransactionRepository: NewUserBalanceTransactionRepository(res),
		RewardRepository:                 NewRewardRepository(res),
		UserRewardRepository:             NewUserRewardRepository(res),
	}
}
