package config

import "time"

type WorkerConfig struct {
	PoolSize              int           `mapstructure:"pool_size"`
	HealthMonitorInterval time.Duration `mapstructure:"health_monitor_interval"`
}
