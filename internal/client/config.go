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
	sslCaCert     string
}

func NewConfig(configPath string) (*Config, error) {
	globalConfig, _ := config.NewConfig(configPath)
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configPath)
	v.ReadInConfig()

	cfg := &Config{
		userID:        globalConfig.UserID,
		datasetID:     globalConfig.DatasetFolder,
		datasetFolder: globalConfig.DatasetFolder,
		sslCaCert:     globalConfig.SslCaCert,
		apiHost:       v.GetString("CLIENT_API_HOST"),
		accessToken:   v.GetString("CLIENT_ACCESS_TOKEN"),
	}

	return cfg, nil
}
