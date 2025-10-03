package job

import (
	"fmt"
)

type Type string

const (
	InitClaim       Type = "init_claim"
	CompleteClaim   Type = "complete_claim"
	KYCVerification Type = "kyc_verification"
)

func (s *Type) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan JobType from %T", value)
	}
	*s = Type(str)
	return nil
}
