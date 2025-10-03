package config

type RouterConfig struct {
	AllowedOrigins string `mapstructure:"allowed_origins"`
	AllowedHeaders string `mapstructure:"allowed_headers"`
}
