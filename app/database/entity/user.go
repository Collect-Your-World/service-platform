package entity

import (
	"backend/service-platform/app/database/constant/role"
	"backend/service-platform/app/database/constant/user"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID            uuid.UUID   `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Username      string      `bun:"username,notnull,unique"`
	Email         *string     `bun:"email,unique"`
	PhoneNumber   *string     `bun:"phone_number,unique"`
	Password      string      `bun:"password,notnull"`
	Status        user.Status `bun:"status,notnull,default:'UNVERIFIED'"`
	Role          role.Role   `bun:"role,notnull,default:'USER'"`
	EmailVerified bool        `bun:"email_verified,default:false"`
	PhoneVerified bool        `bun:"phone_verified,default:false"`
	LastLoginAt   *time.Time  `bun:"last_login_at"`
	CreatedAt     time.Time   `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     *time.Time  `bun:"updated_at"`
	DeletedAt     *time.Time  `bun:"deleted_at,soft_delete"`
	DeactivatedAt *time.Time  `bun:"deactivated_at"`
}

func (u User) Alias() string {
	return "u"
}
