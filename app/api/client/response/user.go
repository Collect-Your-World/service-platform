package response

import (
	"github.com/google/uuid"
)

type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email *string   `json:"email"`
}
