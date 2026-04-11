package managers

import (
	authmanagers "backend/service-platform/app/internal/auth/managers"
	collmanagers "backend/service-platform/app/internal/collection/managers"
	jobmanagers "backend/service-platform/app/internal/job/managers"
	"backend/service-platform/app/internal/platform/runtime"
	"backend/service-platform/app/internal/repository"
	usermanagers "backend/service-platform/app/internal/user/managers"
	"backend/service-platform/app/pkg/bcrypt"
	"backend/service-platform/app/pkg/jwt"
	"backend/service-platform/app/pkg/queue"
)

type Managers struct {
	AuthManager         authmanagers.AuthManager
	JobManager          jobmanagers.JobManager
	UserBalanceManager  usermanagers.UserBalanceManager
	CollectionManager   collmanagers.CollectionManager
	ItemManager         collmanagers.ItemManager
	RarityConfigManager collmanagers.RarityConfigManager
}

func NewManagers(
	res runtime.Resource,
	_ interface{},
	repositories *repository.Repositories,
) *Managers {
	bcryptHasher := bcrypt.NewBcrypt(res.Config.BcryptConfig.Cost)
	hasher := &bcryptHasher

	jwtManager := jwt.NewJwt(res.Config.JwtConfig)

	redisQueue := queue.NewRedisQueue(res.Redis.GetUniversalClient(), res.Logger)
	jobManager := jobmanagers.NewJobManager(repositories.JobRepository, redisQueue, res.Logger)

	return &Managers{
		AuthManager:         authmanagers.NewAuthManager(res, hasher, jwtManager, repositories),
		JobManager:          jobManager,
		UserBalanceManager:  usermanagers.NewUserBalanceManager(repositories.UserBalanceRepository, repositories.UserBalanceTransactionRepository),
		CollectionManager:   collmanagers.NewCollectionManager(res.DB, repositories.CollectionRepository, repositories.CollectionItemRepository),
		ItemManager:         collmanagers.NewItemManager(res.DB, repositories.ItemRepository),
		RarityConfigManager: collmanagers.NewRarityConfigManager(repositories.RarityConfigRepository),
	}
}
