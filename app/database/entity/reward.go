package entity

import (
	rwconst "backend/service-platform/app/database/constant/reward"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type Reward struct {
	bun.BaseModel `bun:"table:rewards,alias:r"`

	ID          uuid.UUID               `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Code        string                  `bun:"code,notnull,unique"`
	Type        rwconst.RewardType      `bun:"type,notnull"`
	Name        string                  `bun:"name,notnull"`
	Description *string                 `bun:"description,nullzero"`
	ValueInt    *int                    `bun:"value_int"`
	MetaJSONB   *map[string]interface{} `bun:"meta_jsonb,type:jsonb,nullzero"`
	CreatedAt   time.Time               `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt   *time.Time              `bun:"updated_at"`
	DeletedAt   *time.Time              `bun:"deleted_at,soft_delete"`
}

func (Reward) Alias() string {
	return "r"
}
