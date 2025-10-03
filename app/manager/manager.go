package manager

import (
	"backend/service-platform/app/database/repository"
	"backend/service-platform/app/internal/runtime"
	"backend/service-platform/app/pkg/bcrypt"
	"backend/service-platform/app/pkg/jwt"
	"backend/service-platform/app/pkg/queue"
)

type Managers struct {
	AuthManager AuthManager
	JobManager  JobManager
}

func NewManagers(
	res runtime.Resource,
	_ interface{},
	repositories *repository.Repositories,
) *Managers {
	// Create bcrypt hasher from configuration
	bcryptHasher := bcrypt.NewBcrypt(res.Config.BcryptConfig.Cost)
	hasher := &bcryptHasher

	// Create a JWT manager from configuration
	jwtManager := jwt.NewJwt(res.Config.JwtConfig)

	// Initialize job-related components
	redisQueue := queue.NewRedisQueue(res.Redis.GetUniversalClient(), res.Logger)
	jobManager := NewJobManager(repositories.JobRepository, redisQueue, res.Logger)

	return &Managers{
		AuthManager: NewAuthManager(res, hasher, jwtManager, repositories),
		JobManager:  jobManager,
	}
}
