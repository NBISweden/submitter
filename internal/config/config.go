package config

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

type Config struct {
	DatasetFolder string `mapstructure:"DATASET_FOLDER"`
	DatasetID     string `mapstructure:"DATASET_ID"`
	UserID        string `mapstructure:"USER_ID"`
	SslCaCert     string `mapstructure:"SSL_CA_CERT"`
	Timeout       int    `mapstructure:"JOB_TIMEOUT"`
	PollRate      int    `mapstructure:"JOB_POLL_RATE"`
}

func NewConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		slog.Info("no configuration file found, continuing configuration with environment variables")
	}

	v.SetDefault("JOB_TIMEOUT", 4320)
	v.SetDefault("JOB_POLL_RATE", 180)

	cfg := &Config{
		DatasetFolder: v.GetString("DATASET_FOLDER"),
		DatasetID:     v.GetString("DATASET_ID"),
		UserID:        v.GetString("USER_ID"),
		SslCaCert:     v.GetString("SSL_CA_CERT"),
	}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.DatasetFolder == "" {
		return fmt.Errorf("DATASET_FOLDER requiered")
	}

	if cfg.DatasetID == "" {
		return fmt.Errorf("DATASET_ID requiered")
	}

	if cfg.UserID == "" {
		return fmt.Errorf("USER_ID requiered")
	}

	if cfg.PollRate > cfg.Timeout {
		return fmt.Errorf("JOB_POLL_RATE greater than JOB_TIMEOUT, set a pollrate that is less than the timeout value")
	}
	return nil
}
