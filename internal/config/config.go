package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	UserID        string
	Uploader      string
	UploaderEmail string
	DatasetID     string
	DatasetFolder string
	Email         string
	Password      string
	APIHost       string
	SMTPHost      string
	SMTPPort      int
	AccessToken   string
	UseTLS        bool
	SSLCACert     string
}

func NewConfig() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	cfg := &Config{
		UserID:        v.GetString("USER_ID"),
		Uploader:      v.GetString("UPLOADER"),
		UploaderEmail: v.GetString("UPLOADER_EMAIL"),
		DatasetID:     v.GetString("DATASET_ID"),
		DatasetFolder: v.GetString("DATASET_FOLDER"),
		Email:         v.GetString("EMAIL"),
		Password:      v.GetString("PASSWORD"),
		APIHost:       v.GetString("API_HOST"),
		SMTPHost:      v.GetString("SMTP_HOST"),
		SMTPPort:      v.GetInt("SMTP_PORT"),
		AccessToken:   v.GetString("ACCESS_TOKEN"),
		UseTLS:        v.GetBool("USE_TLS"),
		SSLCACert:     v.GetString("SSL_CA_CERT"),
	}

	return cfg, nil
}
