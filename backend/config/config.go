package config

import (
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

	fileName string `envconfig:"CONFIG_FILE"`
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

type LogLevel log.Lvl

func (lvl *LogLevel) Lvl() log.Lvl {
	return log.Lvl(*lvl)
}

func (lvl *LogLevel) UnmarshalYAML(node *yaml.Node) error {
	value := strings.ToUpper(node.Value)
	switch value {
	case "DEBUG":
		*lvl = LogLevel(log.DEBUG)
		return nil
	case "INFO":
		*lvl = LogLevel(log.INFO)
		return nil
	case "WARN":
		*lvl = LogLevel(log.WARN)
		return nil
	case "ERROR":
		*lvl = LogLevel(log.ERROR)
		return nil
	case "OFF":
		*lvl = LogLevel(log.OFF)
		return nil
	}

	return errors.New(fmt.Sprintf("unknown log level '%s'", node.Value))
}

func Default() *Config {
	return &Config{
		App: AppConfig{
			Port: 8080,
			DB: DBConfig{
				URL: "",
			},
			Security: SecurityConfig{
				JwtSecret: "",
			},
			Logger: LoggerConfig{
				Level: LogLevel(log.INFO),
			},
		},

		fileName: "config.yaml",
	}
}

func LoadConfig(logger echo.Logger) (*Config, error) {
	cfg := Default()

	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	if err := readFile(cfg); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		logger.Error("Configuration file not found, keeping defaults")
	}
	if err := checkRequiredConfigPresent(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func readFile(cfg *Config) error {
	f, err := os.Open(cfg.fileName)
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

func checkRequiredConfigPresent(cfg *Config) error {
	if cfg.App.DB.URL == "" {
		return errors.New("database URL is empty")
	}
	if cfg.App.Security.JwtSecret == "" {
		return errors.New("JWT secret is empty")
	}

	return nil
}
