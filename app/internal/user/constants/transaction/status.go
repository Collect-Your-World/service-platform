package transaction

type Status string

const (
	PENDING   Status = "PENDING"
	COMPLETED Status = "COMPLETED"
	FAILED    Status = "FAILED"
)
