package config

import "github.com/spf13/viper"

type Conf struct {
	WebServerHost        string `mapstructure:"WEB_SERVER_HOST"`
	WebServerPort        string `mapstructure:"WEB_SERVER_PORT"`
	RedisHost            string `mapstructure:"REDIS_HOST"`
	RedisPort            string `mapstructure:"REDIS_PORT"`
	RedisReadTimeout     int    `mapstructure:"REDIS_READ_TIMEOUT"`
	RedisWriteTimeout    int    `mapstructure:"REDIS_WRITE_TIMEOUT"`
	RateLimitMaxRequests int    `mapstructure:"RATE_LIMIT_MAX_REQUESTS"`
	RateLimitBlockTime   int    `mapstructure:"RATE_LIMIT_BLOCK_TIME"`
}

func LoadConfig(path string) (*Conf, error) {
	var cfg *Conf
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&cfg)

	return cfg, err
}
