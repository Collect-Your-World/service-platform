package job

type EventType string

const (
	ClaimInitiated EventType = "claim.initiated"
	ClaimCompleted EventType = "claim.completed"
)

func (e EventType) String() string {
	return string(e)
}

func (e EventType) ToJobType() Type {
	switch e {
	case ClaimInitiated:
		return InitClaim
	case ClaimCompleted:
		return CompleteClaim
	default:
		return ""
	}
}

func (e EventType) ToPriority() Priority {
	switch e {
	case ClaimInitiated, ClaimCompleted:
		return PriorityHigh
	default:
		return PriorityNormal
	}
}
