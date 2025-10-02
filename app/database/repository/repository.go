package repository

import (
	"backend/service-platform/app/internal/runtime"
)

type Repositories struct {
	UserRepository    UserRepository
	SessionRepository SessionRepository
	JobRepository     JobRepository
}

func NewRepositories(res runtime.Resource) *Repositories {
	return &Repositories{
		UserRepository:    NewUserRepository(res),
		SessionRepository: NewSessionRepository(res),
		JobRepository:     NewJobRepository(res),
	}
}
