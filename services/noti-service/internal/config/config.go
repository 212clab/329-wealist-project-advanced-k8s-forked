package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Auth     AuthConfig     `yaml:"auth"`
	App      AppConfig      `yaml:"app"`
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	Env      string `yaml:"env"`
	LogLevel string `yaml:"log_level"`
}

type DatabaseConfig struct {
	URL string `yaml:"url"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type AuthConfig struct {
	ServiceURL     string `yaml:"service_url"`
	InternalAPIKey string `yaml:"internal_api_key"`
	SecretKey      string `yaml:"secret_key"`
}

type AppConfig struct {
	CacheUnreadTTL int `yaml:"cache_unread_ttl"` // seconds
	CleanupDays    int `yaml:"cleanup_days"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:     8002,
			Env:      "dev",
			LogLevel: "debug",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
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

	// Override with environment variables
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if env := os.Getenv("ENV"); env != "" {
		cfg.Server.Env = env
	}
	if env := os.Getenv("NODE_ENV"); env != "" {
		cfg.Server.Env = env
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Server.LogLevel = logLevel
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.Database.URL = dbURL
	}
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		cfg.Redis.Host = redisHost
	}
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		if p, err := strconv.Atoi(redisPort); err == nil {
			cfg.Redis.Port = p
		}
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		cfg.Redis.Password = redisPassword
	}
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
