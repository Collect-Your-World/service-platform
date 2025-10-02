package config

type RedisConfig struct {
	Hosts           string `mapstructure:"hosts"`
	PoolSize        int    `mapstructure:"pool_size"`
	MinIdleConns    int    `mapstructure:"min_idle_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	WriteTimeout    int    `mapstructure:"write_timeout"`
	ReadTimeout     int    `mapstructure:"read_timeout"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}
