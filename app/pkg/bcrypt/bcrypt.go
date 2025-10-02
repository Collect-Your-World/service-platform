package bcrypt

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Bcrypt struct {
	cost int
}

func NewBcrypt(cost int) Bcrypt {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return Bcrypt{
		cost: cost,
	}
}

func NewBcryptWithDefaultCost() Bcrypt {
	return NewBcrypt(bcrypt.DefaultCost)
}

func (b *Bcrypt) Cost() int {
	return b.cost
}

func (b *Bcrypt) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

func (b *Bcrypt) CheckPassword(password, hash string) (bool, error) {
	if password == "" {
		return false, fmt.Errorf("password cannot be empty")
	}
	if hash == "" {
		return false, fmt.Errorf("hash cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check password: %w", err)
	}

	return true, nil
}
