package util

import (
	"database/sql"
	"errors"

	"github.com/uptrace/bun/driver/pgdriver"
)

func SkipNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func IsUniqueViolation(err error) bool {
	if postgresErr, ok := err.(pgdriver.Error); ok {
		if postgresErr.Field('C') == "23505" { // unique_violation, see at: https://www.postgresql.org/docs/current/errcodes-appendix.html
			return true
		}
	}
	return false
}
