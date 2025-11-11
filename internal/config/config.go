package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

type Config struct {
	UserID        string
	Uploader      string
	UploaderEmail string
	DatasetID     string
	DatasetFolder string
	DataDirectory string
	Email         string
	Password      string
	APIHost       string
	SMTPHost      string
	SMTPPort      int
	AccessToken   string
	UseTLS        bool
	SSLCACert     string
}

func NewConfig(configPath string) (Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetDefault("DATA_DIRECTORY", "data")
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		slog.Info("No config file found, continuing with env vars and defaults", "err", err, "config_file_used", v.ConfigFileUsed())
	}

	cfg := &Config{
		UserID:        v.GetString("USER_ID"),
		Uploader:      v.GetString("UPLOADER"),
		UploaderEmail: v.GetString("UPLOADER_EMAIL"),
		DatasetID:     v.GetString("DATASET_ID"),
		DatasetFolder: v.GetString("DATASET_FOLDER"),
		DataDirectory: v.GetString("DATA_DIRECTORY"),
		Email:         v.GetString("EMAIL"),
		Password:      v.GetString("PASSWORD"),
		APIHost:       v.GetString("API_HOST"),
		SMTPHost:      v.GetString("SMTP_HOST"),
		SMTPPort:      v.GetInt("SMTP_PORT"),
		AccessToken:   v.GetString("ACCESS_TOKEN"),
		UseTLS:        v.GetBool("USE_TLS"),
		SSLCACert:     v.GetString("SSL_CA_CERT"),
	}

	return *cfg, nil
}
