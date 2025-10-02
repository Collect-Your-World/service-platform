package config

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	ctxutil "backend/service-platform/app/pkg/util/context"
)

// bindEnv binds an environment variable with an optional default value
func bindEnv(configKey, envKey string, defaultValue ...interface{}) {
	if len(defaultValue) > 0 {
		viper.SetDefault(configKey, defaultValue[0])
	}
	viper.BindEnv(configKey, envKey)
}

type ApplicationConfig struct {
	ServerConfig     ServerConfig     `mapstructure:"server"`
	DatabaseConfig   DatabaseConfig   `mapstructure:"database"`
	RedisConfig      RedisConfig      `mapstructure:"redis"`
	RouterConfig     RouterConfig     `mapstructure:"router"`
	WorkerConfig     WorkerConfig     `mapstructure:"worker"`
	AwsConfig        AwsConfig        `mapstructure:"aws"`
	EodhdConfig      EodhdConfig      `mapstructure:"eodhd"`
	GoogleConfig     GoogleConfig     `mapstructure:"google"`
	BcryptConfig     BcryptConfig     `mapstructure:"bcrypt"`
	SuperAdminConfig SuperAdminConfig `mapstructure:"super_admin"`
	JwtConfig        JwtConfig        `mapstructure:"jwt"`
}

func ReadApplicationConfig(env ctxutil.AppMode, logger *zap.Logger) (cfg ApplicationConfig, err error) {
	if env == "" {
		env = ctxutil.AppModeLocal
	}
	confFileName := fmt.Sprintf("config-%s", env)

	viper.SetConfigName(confFileName)
	viper.SetConfigType("yaml")

	configPath := "./config"
	if env == ctxutil.AppModeTest {
		configPath = "../../../config"
	}
	viper.AddConfigPath(configPath)
	// For unit tests
	viper.AddConfigPath("../../../../config")

	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("error reading config file: %v", err)
	} else {
		logger.Info(
			"using config",
			zap.String("file", confFileName), // viper.ConfigFileUsed()
		)
	}
	viper.AutomaticEnv()

	// Server
	bindEnv("server.port", "SERVER_PORT")

	// Database
	bindEnv("database.protocol", "DB_PROTOCOL")
	bindEnv("database.url", "DB_URL")
	bindEnv("database.replica_url", "DB_REPLICA_URL")
	bindEnv("database.name", "DB_NAME")
	bindEnv("database.port", "DB_PORT")
	bindEnv("database.username", "DB_USERNAME")
	bindEnv("database.password", "DB_PASSWORD")
	bindEnv("database.ssl_mode", "SSL_MODE")
	bindEnv("database.max_db_conns", "DB_MAX_DB_CONNS")
	bindEnv("database.max_idle_db_conns", "DB_MAX_IDLE_DB_CONNS")
	bindEnv("database.max_conn_lifetime", "DB_MAX_CONN_LIFETIME")
	bindEnv("database.max_conn_idle_time", "DB_MAX_CONN_IDLE_TIME")

	// Redis
	bindEnv("redis.hosts", "REDIS_HOSTS")
	bindEnv("redis.pool_size", "REDIS_POOL_SIZE")
	bindEnv("redis.min_idle_conns", "REDIS_MIN_IDLE_CONNS")
	bindEnv("redis.max_idle_conns", "REDIS_MAX_IDLE_CONNS")
	bindEnv("redis.write_timeout", "REDIS_WRITE_TIMEOUT")
	bindEnv("redis.read_timeout", "REDIS_READ_TIMEOUT")
	bindEnv("redis.conn_max_lifetime", "REDIS_CONN_MAX_LIFETIME")

	// AWS
	bindEnv("aws.region", "AWS_REGION")
	bindEnv("aws.endpoint", "AWS_ENDPOINT")
	bindEnv("aws.sqs.queue_urls.sqs_scheduled_job_queue", "AWS_SQS_SCHEDULED_JOB_QUEUE_URL")
	bindEnv("aws.sqs.polling.max_messages", "AWS_SQS_MAX_MESSAGES")
	bindEnv("aws.sqs.polling.wait_time_seconds", "AWS_SQS_WAIT_TIME_SECONDS")
	bindEnv("aws.sqs.polling.visibility_timeout_seconds", "AWS_SQS_VISIBILITY_TIMEOUT_SECONDS")
	bindEnv("aws.sqs.polling.polling_interval", "AWS_SQS_POLLING_INTERVAL")
	bindEnv("aws.sqs.message.max_retries", "AWS_SQS_MAX_RETRIES")
	bindEnv("aws.sqs.message.base_retry_delay", "AWS_SQS_BASE_RETRY_DELAY")
	bindEnv("aws.sqs.message.max_retry_delay", "AWS_SQS_MAX_RETRY_DELAY")

	// Worker
	bindEnv("worker.pool_size", "WORKER_POLL_SIZE", 2)
	bindEnv("worker.health_monitor_interval", "WORKER_HEALTH_MONITOR_INTERVAL", "2m")

	// Router
	bindEnv("router.allowed_origins", "ROUTER_ALLOWED_ORIGINS")
	bindEnv("router.allowed_headers", "ROUTER_ALLOWED_HEADERS")

	// Google
	bindEnv("google.spreadsheet_id", "GOOGLE_SPREADSHEET_ID")
	bindEnv("google.credentials_file_path", "GOOGLE_CREDENTIALS_FILE_PATH")
	bindEnv("google.credentials_file_json", "GOOGLE_CREDENTIALS_FILE_JSON")

	// TRM
	bindEnv("trm.url", "TRM_URL")
	bindEnv("trm.organization_api_key", "TRM_ORGANIZATION_API_KEY")

	// KYC
	bindEnv("kyc.sanctioned_blockchain_addresses", "KYC_SANCTIONED_BLOCKCHAIN_ADDRESSES")

	// Financial
	bindEnv("financial_service.url", "WHITELIST_URL")
	bindEnv("financial_service.use_mock", "WHITELIST_USE_MOCK")
	bindEnv("financial_service.invalid_addresses", "WHITELIST_INVALID_ADDRESSES")

	// Bcrypt
	bindEnv("bcrypt.cost", "BCRYPT_COST")

	// Super Admin
	bindEnv("super_admin.allowed_new_creation", "SUPER_ADMIN_ALLOWED_NEW_CREATION")

	// JWT
	bindEnv("jwt.issuer", "JWT_ISSUER")
	bindEnv("jwt.secret_key", "JWT_SECRET_KEY")
	bindEnv("jwt.access_expiration", "JWT_ACCESS_EXPIRATION")
	bindEnv("jwt.refresh_expiration", "JWT_REFRESH_EXPIRATION")
	bindEnv("siwe.statement", "SIWE_STATEMENT")
	bindEnv("siwe.nonce_ttl", "SIWE_NONCE_TTL")
	bindEnv("siwe.allowed_origins", "SIWE_ALLOWED_ORIGINS")
	bindEnv("siwe.require_chain_id", "SIWE_REQUIRE_CHAIN_ID")

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("error unmarshalling config: %s", err.Error())
	}

	return cfg, err
}
