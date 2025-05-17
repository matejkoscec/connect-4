package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type Config struct {
	App AppConfig
}

type AppConfig struct {
	Port     uint
	DB       DBConfig
	Security SecurityConfig
	Logger   LoggerConfig
}

type DBConfig struct {
	URL string `envconfig:"DB_URL"`
}

type SecurityConfig struct {
	JwtSecret string `envconfig:"JWT_SECRET" yaml:"jwtSecret"`
}

type LoggerConfig struct {
	Level LogLevel
}

type LogLevel struct {
	log.Lvl
}

const DefaultFileName = "config.yaml"

var DefaultConfig = &Config{
	App: AppConfig{
		Port: 8080,
		DB: DBConfig{
			URL: "",
		},
		Security: SecurityConfig{
			JwtSecret: "",
		},
		Logger: LoggerConfig{
			Level: LogLevel{log.INFO},
		},
	},
}

func LoadConfig(logger echo.Logger) (*Config, error) {
	cfg := DefaultConfig

	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	if err := cfg.loadFile(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		logger.Error("Configuration file not found, keeping defaults")
	}

	return cfg, nil
}

func (cfg *Config) loadFile() error {
	f, err := os.Open(DefaultFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *Config) PrettyString() string {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

func (lvl LogLevel) UnmarshalYAML(node *yaml.Node) error {
	switch strings.ToUpper(node.Value) {
	case "DEBUG":
		lvl.Lvl = log.DEBUG
		return nil
	case "INFO":
		lvl.Lvl = log.INFO
		return nil
	case "WARN":
		lvl.Lvl = log.WARN
		return nil
	case "ERROR":
		lvl.Lvl = log.ERROR
		return nil
	case "OFF":
		lvl.Lvl = log.OFF
		return nil
	}

	return errors.New(fmt.Sprintf("unknown log level '%s'", node.Value))
}
