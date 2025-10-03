package request

type FindEmailRequest struct {
	Email string `query:"email" validate:"required,email"`
}

type AuthRequest struct {
	Email string `json:"email" validate:"required,email"`
	UUID  string `json:"uuid" validate:"required,uuid4"`
}
