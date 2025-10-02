package config

type EodhdConfig struct {
	BaseAPI string `mapstructure:"base_api"`
	Token   string `mapstructure:"token"`
}
