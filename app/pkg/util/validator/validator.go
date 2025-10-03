package validator

import (
	"fmt"
	"net/url"
	"strings"

	"backend/service-platform/app/pkg/util/constants"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

func ValidateJobType(jobType string) error {
	validJobTypes := []string{
		constants.EngineTypeUnspecified,
		constants.EngineTypeCustom,
		constants.EngineTypeMPI,
		constants.EngineTypeSpark,
		constants.EngineTypeMesa,
		constants.EngineTypeAthena,
		constants.EngineTypeLLM,
		constants.EngineTypeESO,
		constants.EngineTypeChanga,
		constants.EngineTypeNextflow,
		constants.EngineTypeStashTransfer,
		constants.EngineTypeGromacs,
		constants.EngineTypeLammps,
		constants.EngineTypeOpenfoam,
	}

	for _, validType := range validJobTypes {
		if jobType == validType {
			return nil
		}
	}

	return ValidationError{
		Field:   "job_type",
		Message: fmt.Sprintf("must be one of: %s", strings.Join(validJobTypes, ", ")),
	}
}

func ValidateEngineSize(engineSize string) error {
	validEngineSizes := []string{
		constants.EngineSizeUnspecified,
		constants.EngineSizeXMicro,
		constants.EngineSizeMicro,
		constants.EngineSizeXxsmall,
		constants.EngineSizeXsmall,
		constants.EngineSizeSmall,
		constants.EngineSizeMedium,
		constants.EngineSizeLarge,
		constants.EngineSizeXlarge,
		constants.EngineSizeXXlarge,
	}

	for _, validSize := range validEngineSizes {
		if engineSize == validSize {
			return nil
		}
	}

	return ValidationError{
		Field:   "engine_size",
		Message: fmt.Sprintf("must be one of: %s", strings.Join(validEngineSizes, ", ")),
	}
}

func ValidateNumEngines(numEngines int) error {
	if numEngines < 1 {
		return ValidationError{
			Field:   "num_engines",
			Message: "must be at least 1",
		}
	}
	return nil
}

func ValidateGPUEngineSize(engineSize string) error {
	supportedGPUEngineSizes := []string{
		constants.EngineSizeXsmall,
		constants.EngineSizeMedium,
		constants.EngineSizeLarge,
		constants.EngineSizeXlarge,
		constants.EngineSizeXXlarge,
	}

	for _, validSize := range supportedGPUEngineSizes {
		if engineSize == validSize {
			return nil
		}
	}

	return ValidationError{
		Field:   "engine_size",
		Message: fmt.Sprintf("unsupported GPU worker type, must be one of: %s", strings.Join(supportedGPUEngineSizes, ", ")),
	}
}

func ValidateGPUJobType(jobType string) error {
	supportedGPUJobTypes := []string{
		constants.EngineTypeLammps,
		constants.EngineTypeGromacs,
		constants.EngineTypeMPI,
		constants.EngineTypeAthena,
		constants.EngineTypeLLM,
	}

	for _, validType := range supportedGPUJobTypes {
		if jobType == validType {
			return nil
		}
	}

	return ValidationError{
		Field:   "job_type",
		Message: fmt.Sprintf("job type %s is not supported with GPU", jobType),
	}
}

func ValidateRequiredString(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{
			Field:   fieldName,
			Message: "is required",
		}
	}
	return nil
}

func ValidateEngineSizeSet(workerType, engineSize *string) error {
	if (workerType == nil || *workerType == "") && (engineSize == nil || *engineSize == "") {
		return ValidationError{
			Field:   "engine_size",
			Message: "engine_size must be set",
		}
	}
	if engineSize == nil || *engineSize == "" {
		return ValidationError{
			Field:   "engine_size",
			Message: "engine_size must be set, you might be on an older client, please restart your notebook server by relogging",
		}
	}
	return nil
}

func ValidateSupportedVersion(jobType string, imageTag *string) error {
	config, ok := constants.SupportedJobConfigs[jobType]
	if !ok || len(config.SupportedVersions) == 0 || imageTag == nil || *imageTag == "" {
		return nil
	}
	for _, v := range config.SupportedVersions {
		if *imageTag == v {
			return nil
		}
	}
	return ValidationError{
		Field:   "image_tag",
		Message: "unsupported version for job type " + jobType,
	}
}

func IsDomainAllowed(domain string, allowed []string) bool {
	if domain == "" || len(allowed) == 0 {
		return false
	}
	d := strings.ToLower(strings.TrimSpace(domain))
	for _, raw := range allowed {
		a := strings.TrimSpace(raw)
		if a == "" {
			continue
		}
		if hasScheme(a) {
			if u, err := url.Parse(a); err == nil {
				a = u.Hostname()
			}
		}
		a = strings.ToLower(a)

		if strings.HasPrefix(a, "*.") {
			suffix := strings.TrimPrefix(a, "*.")
			if d == suffix || strings.HasSuffix(d, "."+suffix) {
				return true
			}
			continue
		}
		if strings.HasPrefix(a, ".") {
			suffix := strings.TrimPrefix(a, ".")
			if d == suffix || strings.HasSuffix(d, "."+suffix) {
				return true
			}
			continue
		}
		if d == a {
			return true
		}
	}
	return false
}

func hasScheme(s string) bool {
	i := strings.Index(s, "://")
	if i <= 0 {
		return false
	}
	for j := 0; j < i; j++ {
		c := s[j]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}
	return true
}
