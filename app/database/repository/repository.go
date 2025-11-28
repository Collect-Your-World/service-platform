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
	CollectionRepository             CollectionRepository
	CollectionItemRepository         CollectionItemRepository
	ItemRepository                   ItemRepository
	RarityConfigRepository           RarityConfigRepository
}

func NewRepositories(res runtime.Resource) *Repositories {
	return &Repositories{
		UserRepository:                   NewUserRepository(res),
		SessionRepository:                NewSessionRepository(res),
		JobRepository:                    NewJobRepository(res),
		UserBalanceRepository:            NewUserBalanceRepository(res),
		UserBalanceTransactionRepository: NewUserBalanceTransactionRepository(res),
		CollectionRepository:             NewCollectionRepository(res),
		CollectionItemRepository:         NewCollectionItemRepository(res),
		ItemRepository:                   NewItemRepository(res),
		RarityConfigRepository:           NewRarityConfigRepository(res),
	}
}
