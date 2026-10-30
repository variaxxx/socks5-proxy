package config

import (
	"bytes"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Auth AuthConfig `yaml:"auth"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int `yaml:"port"`
}

type AuthConfig struct {
	RequireAuth bool `yaml:"require_auth"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 1080,
		},
		Auth: AuthConfig{
			RequireAuth: false,
			Username: "",
			Password: "",
		},
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", cfg.Server.Port)
	}

	if cfg.Server.Host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}

	if cfg.Auth.RequireAuth {
		if cfg.Auth.Username == "" || cfg.Auth.Password == "" {
			return nil, fmt.Errorf("invalid username or password")
		}
		if len([]byte(cfg.Auth.Username)) > 255 || len([]byte(cfg.Auth.Password)) > 255 {
			return nil, fmt.Errorf("too long username or password")
		}
	}

	return cfg, nil
}