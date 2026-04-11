package worker

type Kind string

const (
	SocialsSynchronizationJobKind   Kind = "socials-synchronization-job"
	CompaniesSynchronizationJobKind Kind = "companies-synchronization-job"
)

type Payload struct {
	Kind Kind `json:"kind"`
}
