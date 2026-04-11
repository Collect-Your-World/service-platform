package role

import (
	"database/sql/driver"
	"fmt"
)

// Role represents a role in the role-based access control system
type Role string

const (
	// User represents a regular user with basic permissions
	User Role = "USER"
	// Admin represents an administrator with elevated permissions
	Admin Role = "ADMIN"
	// Operator represents an operator with specific operational permissions
	Operator Role = "OPERATOR"
	// SuperAdmin represents a super administrator with full system access
	SuperAdmin Role = "SUPER_ADMIN"
)

// Scan implements the sql.Scanner interface for database scanning
func (r *Role) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan Role from %T", value)
	}
	*r = Role(str)
	return nil
}

// Value implements the driver.Valuer interface for database storage
func (r Role) Value() (driver.Value, error) {
	return string(r), nil
}
