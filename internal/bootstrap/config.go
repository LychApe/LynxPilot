package bootstrap

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Auth   AuthConfig   `yaml:"auth"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type AuthConfig struct {
	TokenSalt string `yaml:"token_salt"`
}

func ResolveConfigPath() string {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return "config/config.yaml"
	}
	return configPath
}

func LoadConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{Port: 8080},
	}
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, err
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	return cfg, nil
}

func (c *Config) HTTPAddr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}
