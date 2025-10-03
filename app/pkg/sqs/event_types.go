package sqs

// EventType represents the types of events that can be processed via SQS
type EventType string

const (
	Claim           EventType = "claim"
	KYCVerification EventType = "kyc.verification"
)

// String returns the string representation of the event type
func (e EventType) String() string {
	return string(e)
}

// IsValid checks if the event type is valid
func (e EventType) IsValid() bool {
	switch e {
	case Claim, KYCVerification:
		return true
	default:
		return false
	}
}

// GetAllEventTypes returns all valid event types
func GetAllEventTypes() []EventType {
	return []EventType{
		Claim,
		KYCVerification,
	}
}
