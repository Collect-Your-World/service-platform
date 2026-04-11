package job

import (
	"fmt"
)

type Status string

const (
	Pending    Status = "pending"
	Processing Status = "processing"
	Completed  Status = "completed"
	Failed     Status = "failed"
	Retrying   Status = "retrying"
)

func (s *Status) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan JobStatus from %T", value)
	}
	*s = Status(str)
	return nil
}
