package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserReward struct {
	bun.BaseModel `bun:"table:user_rewards,alias:ur"`

	ID        uuid.UUID  `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID    uuid.UUID  `bun:"user_id,notnull"`
	RewardID  uuid.UUID  `bun:"reward_id,notnull"`
	Reason    *string    `bun:"reason,nullzero"` // e.g., "leaderboard_top10", "daily_login"
	GrantedAt time.Time  `bun:"granted_at,notnull,default:current_timestamp"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt *time.Time `bun:"updated_at"`
	DeletedAt *time.Time `bun:"deleted_at,soft_delete"`
}

func (UserReward) Alias() string {
	return "ur"
}
