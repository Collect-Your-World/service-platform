package sqs

// SqsMessage represents the base structure for all SQS messages
type SqsMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ClaimPayload represents the payload for claim events
type ClaimPayload struct {
	User   string  `json:"user"`
	Amount float64 `json:"amount,omitempty"`
}

// KYCVerificationPayload represents the payload for KYC verification events
type KYCVerificationPayload struct {
	User     string `json:"user"`
	UserID   string `json:"user_id,omitempty"`
	Document string `json:"document,omitempty"`
}

// GetPayloadForEventType returns the appropriate payload struct for the given event type
func GetPayloadForEventType(eventType EventType) interface{} {
	switch eventType {
	case Claim:
		return &ClaimPayload{}
	case KYCVerification:
		return &KYCVerificationPayload{}
	default:
		return &map[string]interface{}{}
	}
}
