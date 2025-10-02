package entity

import (
	"backend/service-platform/app/database/constant/job"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type JobPayload map[string]interface{}

func (p JobPayload) Value() (driver.Value, error) {
	if p == nil {
		return "{}", nil
	}
	data, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JobPayload: %w", err)
	}
	return string(data), nil
}

func (p *JobPayload) Scan(value interface{}) error {
	if value == nil {
		*p = make(JobPayload)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JobPayload", value)
	}

	if len(bytes) == 0 {
		*p = make(JobPayload)
		return nil
	}

	return json.Unmarshal(bytes, p)
}

type Job struct {
	bun.BaseModel `bun:"table:jobs,alias:j"`

	ID          uuid.UUID    `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	Type        string       `bun:"type,notnull" json:"type"`
	Priority    job.Priority `bun:"priority,notnull" json:"priority"`
	Payload     JobPayload   `bun:"payload,type:jsonb" json:"payload"`
	Attempts    int          `bun:"attempts,notnull,default:0" json:"attempts"`
	MaxAttempts int          `bun:"max_attempts,notnull,default:3" json:"max_attempts"`
	CreatedAt   time.Time    `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   *time.Time   `bun:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time   `bun:"deleted_at,nullzero" json:"deleted_at,omitempty"`
	ScheduledAt *time.Time   `bun:"scheduled_at,nullzero" json:"scheduled_at,omitempty"`
	StartedAt   *time.Time   `bun:"started_at,nullzero" json:"started_at,omitempty"`
	CompletedAt *time.Time   `bun:"completed_at,nullzero" json:"completed_at,omitempty"`
	Status      job.Status   `bun:"status,notnull,default:'pending'" json:"status"`
	Error       string       `bun:"error" json:"error,omitempty"`
}
