package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Auth     AuthConfig     `yaml:"auth"`
	UserAPI  UserAPIConfig  `yaml:"user_api"`
	Services ServicesConfig `yaml:"services"`
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	BasePath string `yaml:"base_path"`
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
	ServiceURL string `yaml:"service_url"`
	SecretKey  string `yaml:"secret_key"`
}

type ServicesConfig struct {
	UserServiceURL string `yaml:"user_service_url"`
}

type UserAPIConfig struct {
	BaseURL string        `yaml:"base_url"`
	Timeout time.Duration `yaml:"timeout"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:     8001,
			BasePath: "/api/chats",
			Env:      "dev",
			LogLevel: "debug",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
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
	if basePath := os.Getenv("SERVER_BASE_PATH"); basePath != "" {
		cfg.Server.BasePath = basePath
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
	if secretKey := os.Getenv("SECRET_KEY"); secretKey != "" {
		cfg.Auth.SecretKey = secretKey
	}
	if userURL := os.Getenv("USER_SERVICE_URL"); userURL != "" {
		cfg.Services.UserServiceURL = userURL
		cfg.UserAPI.BaseURL = userURL
	}
	if timeout := os.Getenv("USER_API_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			cfg.UserAPI.Timeout = d
		}
	}
	if cfg.UserAPI.Timeout == 0 {
		cfg.UserAPI.Timeout = 5 * time.Second
	}

	return cfg, nil
}
