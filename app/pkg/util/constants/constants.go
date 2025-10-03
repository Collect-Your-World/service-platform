package constants

const (
	JobStatusUnspecified  = "UNSPECIFIED"
	JobStatusPending      = "PENDING"
	JobStatusRunning      = "RUNNING"
	JobStatusCompleted    = "COMPLETED"
	JobStatusFailed       = "FAILED"
	JobStatusCancelled    = "CANCELLED"
	JobStatusProvisioning = "PROVISIONING"
	JobStatusInitializing = "INITIALIZING"
)

const (
	EngineTypeUnspecified   = "UNSPECIFIED"
	EngineTypeCustom        = "CUSTOM"
	EngineTypeMPI           = "MPI"
	EngineTypeSpark         = "SPARK"
	EngineTypeMesa          = "MESA"
	EngineTypeAthena        = "ATHENA"
	EngineTypeLLM           = "LLM"
	EngineTypeESO           = "ESO"
	EngineTypeChanga        = "CHANGA"
	EngineTypeNextflow      = "NEXTFLOW"
	EngineTypeStashTransfer = "STASH_TRANSFER"
	EngineTypeGromacs       = "GROMACS"
	EngineTypeLammps        = "LAMMPS"
	EngineTypeOpenfoam      = "OPENFOAM"
)

const (
	EngineSizeUnspecified = "UNSPECIFIED"
	EngineSizeXMicro      = "XMICRO"
	EngineSizeMicro       = "MICRO"
	EngineSizeXxsmall     = "XXSMALL"
	EngineSizeXsmall      = "XSMALL"
	EngineSizeSmall       = "SMALL"
	EngineSizeMedium      = "MEDIUM"
	EngineSizeLarge       = "LARGE"
	EngineSizeXlarge      = "XLARGE"
	EngineSizeXXlarge     = "XXLARGE"
)

const (
	ControllerKindJob = "Job"

	ObjectKindJob = "Job"
	ObjectKindPod = "Pod"

	ReasonStarted              = "Started"
	ReasonKilling              = "Killing"
	ReasonCompleted            = "Completed"
	ReasonBackoffLimitExceeded = "BackoffLimitExceeded"
)

type JobConfig struct {
	SupportedVersions []string
}

var SupportedJobConfigs = map[string]JobConfig{
	EngineTypeMesa: {
		SupportedVersions: []string{"r24.03.1", "r23.05.1"},
	},
}
