package config

import "time"

type JwtConfig struct {
	Issuer            string        `mapstructure:"issuer"`
	SecretKey         string        `mapstructure:"secret_key"`
	AccessExpiration  time.Duration `mapstructure:"access_expiration"`
	RefreshExpiration time.Duration `mapstructure:"refresh_expiration"`
}
