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
	LiveKit  LiveKitConfig  `yaml:"livekit"`
	Services ServicesConfig `yaml:"services"`
	CORS     CORSConfig     `yaml:"cors"`
}

type CORSConfig struct {
	AllowedOrigins string `yaml:"allowed_origins"`
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

type LiveKitConfig struct {
	Host      string `yaml:"host"`
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
	WSUrl     string `yaml:"ws_url"`
}

type ServicesConfig struct {
	UserServiceURL string `yaml:"user_service_url"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:     8004,
			BasePath: "/api/video",
			Env:      "dev",
			LogLevel: "debug",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		LiveKit: LiveKitConfig{
			Host:  "http://localhost:7880",
			WSUrl: "ws://localhost:7880",
		},
		CORS: CORSConfig{
			AllowedOrigins: "*",
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
	}

	// LiveKit configuration
	if lkHost := os.Getenv("LIVEKIT_HOST"); lkHost != "" {
		cfg.LiveKit.Host = lkHost
	}
	if lkAPIKey := os.Getenv("LIVEKIT_API_KEY"); lkAPIKey != "" {
		cfg.LiveKit.APIKey = lkAPIKey
	}
	if lkAPISecret := os.Getenv("LIVEKIT_API_SECRET"); lkAPISecret != "" {
		cfg.LiveKit.APISecret = lkAPISecret
	}
	if lkWSUrl := os.Getenv("LIVEKIT_WS_URL"); lkWSUrl != "" {
		cfg.LiveKit.WSUrl = lkWSUrl
	}

	// CORS configuration
	if corsOrigins := os.Getenv("CORS_ORIGINS"); corsOrigins != "" {
		cfg.CORS.AllowedOrigins = corsOrigins
	}

	return cfg, nil
}
