package client

import (
	"github.com/NBISweden/submitter/internal/config"
	"github.com/spf13/viper"
)

type Config struct {
	userID        string
	datasetID     string
	datasetFolder string
	apiHost       string
	accessToken   string
	ssl           bool
	sslCaCert     string
}

func NewConfig(configPath string) (*Config, error) {
	globalConfig, _ := config.NewConfig(configPath)
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configPath)
	v.SetDefault("SSL", false)
	v.ReadInConfig()

	cfg := &Config{
		apiHost:       v.GetString("API_HOST"),
		userID:        globalConfig.UserID,
		datasetID:     globalConfig.DatasetFolder,
		datasetFolder: globalConfig.DatasetFolder,
		accessToken:   v.GetString("ACCESS_TOKEN"),
		ssl:           v.GetBool("SSL"),
		sslCaCert:     v.GetString("SSL_CA_CERT"),
	}

	return cfg, nil
}
