package config

import (
	"fmt"
	"net/url"
)

type DatabaseConfig struct {
	URI             string `mapstructure:"uri"`
	Protocol        string `mapstructure:"protocol"`
	URL             string `mapstructure:"url"`
	ReplicaURL      string `mapstructure:"replica_url"`
	Name            string `mapstructure:"name"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Port            int    `mapstructure:"port"`
	SslMode         string `mapstructure:"ssl_mode"`
	MaxDBConns      int    `mapstructure:"max_db_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_db_conns"`
	MaxConnLifetime int    `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime int    `mapstructure:"max_conn_idle_time"`
}

func (c DatabaseConfig) PrimaryConnectionString() string {
	if c.URI != "" {
		return c.URI
	}
	if c.Username != "" && c.Password != "" {
		return fmt.Sprintf(
			"%s://%s:%s@%s:%d/%s?sslmode=%s",
			c.Protocol, c.Username, url.QueryEscape(c.Password), c.URL, c.Port, c.Name, c.SslMode,
		)
	} else {
		return fmt.Sprintf(
			"%s://%s:%d/%s?sslmode=%s",
			c.Protocol, c.URL, c.Port, c.Name, c.SslMode,
		)
	}
}

func (c DatabaseConfig) ReplicaConnectionString() string {
	if c.Username != "" && c.Password != "" {
		return fmt.Sprintf(
			"%s://%s:%s@%s:%d/%s?sslmode=%s",
			c.Protocol, c.Username, url.QueryEscape(c.Password), c.ReplicaURL, c.Port, c.Name, c.SslMode,
		)
	} else {
		return fmt.Sprintf(
			"%s://%s:%d/%s?sslmode=%s",
			c.Protocol, c.ReplicaURL, c.Port, c.Name, c.SslMode,
		)
	}
}
