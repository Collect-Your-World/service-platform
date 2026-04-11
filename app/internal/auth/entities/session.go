package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID        uuid.UUID  `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID    uuid.UUID  `bun:"user_id,notnull"`
	Token     string     `bun:"token,notnull,unique"`
	UserAgent *string    `bun:"user_agent"`
	IPAddress *string    `bun:"ip_address"`
	Revoked   bool       `bun:"revoked,notnull,default:false"`
	ExpiresAt *time.Time `bun:"expires_at"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt *time.Time `bun:"updated_at"`
	DeletedAt *time.Time `bun:"deleted_at,soft_delete"`
}

func (s Session) Alias() string {
	return "s"
}
