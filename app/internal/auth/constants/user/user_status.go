package user

import (
	"database/sql/driver"
	"fmt"
)

// Status represents the status of a user account
type Status string

const (
	// Verified indicates the user account has been verified
	Verified Status = "VERIFIED"
	// Disabled indicates the user account has been disabled
	Disabled Status = "DISABLED"
	// Unverified indicates the user account has not been verified yet
	Unverified Status = "UNVERIFIED"
)

// Scan implements the sql.Scanner interface for database scanning
func (s *Status) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan UserStatus from %T", value)
	}
	*s = Status(str)
	return nil
}

// Value implements the driver.Valuer interface for database storage
func (s Status) Value() (driver.Value, error) {
	return string(s), nil
}
