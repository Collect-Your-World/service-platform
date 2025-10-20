package transaction

// Status represents the status of transaction
// @Description Transaction status for balance changes
type Status string

const (
	PENDING   Status = "PENDING"
	COMPLETED Status = "COMPLETED"
	FAILED    Status = "FAILED"
)
