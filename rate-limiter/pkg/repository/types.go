package repository

type RateLimitConfig struct {
	Id          *int   `json:"id,omitempty"`
	ConfigValue string `json:"config_value"`
	LimitType   string `json:"limit_type"`
	MaxRequest  int    `json:"max_request"`
	BlockTime   int    `json:"block_time"`
}
