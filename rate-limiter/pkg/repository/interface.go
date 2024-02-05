package repository

type Repository interface {
	GetConfigByID(id string) (RateLimitConfig, error)
	GetConfigByConfigValue(configValue string) (RateLimitConfig, error)
	GetAllConfigs() ([]RateLimitConfig, error)
	CreateConfig(config RateLimitConfig) error
	UpdateConfig(config RateLimitConfig) error
	DeleteConfig(id string) error
}
