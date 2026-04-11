package repository

import (
	authrepo "backend/service-platform/app/internal/auth/repositories"
	collrepo "backend/service-platform/app/internal/collection/repositories"
	jobrepo "backend/service-platform/app/internal/job/repositories"
	"backend/service-platform/app/internal/platform/runtime"
	userrepo "backend/service-platform/app/internal/user/repositories"
)

type Repositories struct {
	UserRepository                   authrepo.UserRepository
	SessionRepository                authrepo.SessionRepository
	JobRepository                    jobrepo.JobRepository
	UserBalanceRepository            userrepo.UserBalanceRepository
	UserBalanceTransactionRepository userrepo.UserBalanceTransactionRepository
	CollectionRepository             collrepo.CollectionRepository
	CollectionItemRepository         collrepo.CollectionItemRepository
	ItemRepository                   collrepo.ItemRepository
	RarityConfigRepository           collrepo.RarityConfigRepository
}

func NewRepositories(res runtime.Resource) *Repositories {
	return &Repositories{
		UserRepository:                   authrepo.NewUserRepository(res),
		SessionRepository:                authrepo.NewSessionRepository(res),
		JobRepository:                    jobrepo.NewJobRepository(res),
		UserBalanceRepository:            userrepo.NewUserBalanceRepository(res),
		UserBalanceTransactionRepository: userrepo.NewUserBalanceTransactionRepository(res),
		CollectionRepository:             collrepo.NewCollectionRepository(res),
		CollectionItemRepository:         collrepo.NewCollectionItemRepository(res),
		ItemRepository:                   collrepo.NewItemRepository(res),
		RarityConfigRepository:           collrepo.NewRarityConfigRepository(res),
	}
}
