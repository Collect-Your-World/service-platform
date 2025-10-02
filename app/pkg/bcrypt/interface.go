package bcrypt

type Hasher interface {
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) (bool, error)
}
