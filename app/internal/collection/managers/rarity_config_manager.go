package managers

import (
	"backend/service-platform/app/internal/collection/entities"
	collrepo "backend/service-platform/app/internal/collection/repositories"
	"context"

	"github.com/google/uuid"
)

type RarityConfigManager interface {
	Create(ctx context.Context, rc *entity.RarityConfig) (*entity.RarityConfig, error)
	Update(ctx context.Context, rc *entity.RarityConfig) (*entity.RarityConfig, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.RarityConfig, error)
	GetByCode(ctx context.Context, code string) (*entity.RarityConfig, error)
	List(ctx context.Context, filter collrepo.RarityConfigFilter) ([]entity.RarityConfig, error)
}

type DefaultRarityConfigManager struct {
	repo collrepo.RarityConfigRepository
}

func NewRarityConfigManager(repo collrepo.RarityConfigRepository) RarityConfigManager {
	return &DefaultRarityConfigManager{repo: repo}
}

func (m *DefaultRarityConfigManager) Create(ctx context.Context, rc *entity.RarityConfig) (*entity.RarityConfig, error) {
	return m.repo.Create(ctx, rc)
}

func (m *DefaultRarityConfigManager) Update(ctx context.Context, rc *entity.RarityConfig) (*entity.RarityConfig, error) {
	return m.repo.Update(ctx, rc)
}

func (m *DefaultRarityConfigManager) GetByID(ctx context.Context, id uuid.UUID) (*entity.RarityConfig, error) {
	return m.repo.FindByID(ctx, id)
}

func (m *DefaultRarityConfigManager) GetByCode(ctx context.Context, code string) (*entity.RarityConfig, error) {
	return m.repo.FindByCode(ctx, code)
}

func (m *DefaultRarityConfigManager) List(ctx context.Context, filter collrepo.RarityConfigFilter) ([]entity.RarityConfig, error) {
	return m.repo.List(ctx, filter)
}
