package config

type AwsConfig struct {
	Region   string    `mapstructure:"region"`
	Endpoint string    `mapstructure:"endpoint"`
	Sqs      SQSConfig `mapstructure:"sqs"`
}

// ToSQSPackageConfig converts the application config to pkg/sqs compatible format
// Note: This method is implemented in the service layer to avoid import cycles
func (c AwsConfig) ToSQSPackageConfig() interface{} {
	return map[string]interface{}{
		"region": c.Region,
		"queue_urls": map[string]string{
			"sqs_scheduled_job_queue": c.Sqs.QueueURLs.SqsScheduledJobQueue,
		},
		"polling": map[string]interface{}{
			"max_messages":               c.Sqs.Polling.MaxMessages,
			"wait_time_seconds":          c.Sqs.Polling.WaitTimeSeconds,
			"visibility_timeout_seconds": c.Sqs.Polling.VisibilityTimeoutSeconds,
			"polling_interval":           c.Sqs.Polling.PollingInterval,
		},
		"message": map[string]interface{}{
			"max_retries":      c.Sqs.Message.MaxRetries,
			"base_retry_delay": c.Sqs.Message.BaseRetryDelay,
			"max_retry_delay":  c.Sqs.Message.MaxRetryDelay,
		},
	}
}
