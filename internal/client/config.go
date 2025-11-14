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
	caCert        string
}

func NewConfig(configPath string) (*Config, error) {
	globalConfig, _ := config.NewConfig(configPath)
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configPath)
	v.SetDefault("SSL", false)
	v.ReadInConfig()

	cfg := &Config{
		userID:        globalConfig.UserID,
		datasetID:     globalConfig.DatasetFolder,
		datasetFolder: globalConfig.DatasetFolder,
		apiHost:       v.GetString("CLIENT_API_HOST"),
		accessToken:   v.GetString("CLIENT_ACCESS_TOKEN"),
		ssl:           v.GetBool("CLIENT_SSL"),
		caCert:        v.GetString("CLIENT_CA_CERT"),
	}

	return cfg, nil
}
