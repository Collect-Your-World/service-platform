package job

import (
	"fmt"
)

type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func (p *Priority) Scan(value interface{}) error {
	if value == nil {
		*p = PriorityNormal
		return nil
	}

	switch v := value.(type) {
	case int64:
		*p = Priority(v)
	case int:
		*p = Priority(v)
	default:
		return fmt.Errorf("cannot scan JobPriority from %T", value)
	}
	return nil
}
