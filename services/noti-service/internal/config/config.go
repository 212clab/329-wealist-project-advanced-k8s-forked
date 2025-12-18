package config

import (
	"os"

	commonconfig "github.com/OrangesCloud/wealist-advanced-go-pkg/config"
	"gopkg.in/yaml.v3"
)

// Config contains all configuration for noti-service.
type Config struct {
	commonconfig.BaseConfig `yaml:",inline"`
	Auth                    NotiAuthConfig `yaml:"auth"`
	App                     AppConfig      `yaml:"app"`
}

// NotiAuthConfig extends the base AuthConfig with noti-specific fields.
type NotiAuthConfig struct {
	ServiceURL     string `yaml:"service_url"`
	InternalAPIKey string `yaml:"internal_api_key"`
	SecretKey      string `yaml:"secret_key"`
}

// AppConfig contains notification-specific configuration.
type AppConfig struct {
	CacheUnreadTTL int `yaml:"cache_unread_ttl"` // seconds
	CleanupDays    int `yaml:"cleanup_days"`
}

// Load reads configuration from yaml file and environment variables.
func Load(path string) (*Config, error) {
	// Start with defaults
	base := commonconfig.DefaultBaseConfig()
	base.Server.Port = 8002

	cfg := &Config{
		BaseConfig: base,
		App: AppConfig{
			CacheUnreadTTL: 300, // 5 minutes
			CleanupDays:    30,
		},
	}

	// Load from yaml file if exists
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	// Override with environment variables (common config)
	cfg.BaseConfig.LoadFromEnv()

	// Override Auth with common BaseConfig values if not set
	if cfg.Auth.ServiceURL == "" {
		cfg.Auth.ServiceURL = cfg.BaseConfig.Auth.ServiceURL
	}
	if cfg.Auth.SecretKey == "" {
		cfg.Auth.SecretKey = cfg.BaseConfig.Auth.SecretKey
	}

	// Service-specific environment variables
	if authURL := os.Getenv("AUTH_SERVICE_URL"); authURL != "" {
		cfg.Auth.ServiceURL = authURL
	}
	if apiKey := os.Getenv("INTERNAL_API_KEY"); apiKey != "" {
		cfg.Auth.InternalAPIKey = apiKey
	}
	if secretKey := os.Getenv("SECRET_KEY"); secretKey != "" {
		cfg.Auth.SecretKey = secretKey
	}

	return cfg, nil
}
