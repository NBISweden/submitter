package client

import (
	"github.com/spf13/viper"
)

type Config struct {
	UserID        string
	DatasetID     string
	DatasetFolder string
	APIHost       string
	AccessToken   string
	SSL           bool
	SSLCACert     string
}

func NewConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configPath)
	v.SetDefault("SSL", false)
	v.ReadInConfig()

	cfg := &Config{
		APIHost:       v.GetString("API_HOST"),
		UserID:        v.GetString("USER_ID"),
		DatasetID:     v.GetString("DATASET_ID"),
		DatasetFolder: v.GetString("DATASET_FOLDER"),
		AccessToken:   v.GetString("ACCESS_TOKEN"),
		SSL:           v.GetBool("SSL"),
		SSLCACert:     v.GetString("SSL_CA_CERT"),
	}

	return cfg, nil
}
