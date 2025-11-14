package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DatasetFolder string
	DatasetID     string
	UserID        string
	SSL           bool
	CaCert        string
}

func NewConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configPath)
	v.ReadInConfig()

	cfg := &Config{
		DatasetFolder: v.GetString("DATASET_FOLDER"),
		DatasetID:     v.GetString("DATASET_ID"),
		UserID:        v.GetString("USER_ID"),
	}

	return cfg, nil
}
