package mail

import (
	"github.com/spf13/viper"
)

type Config struct {
	emailAddress  string
	emailPassword string
	smtpHost      string
	smtpPort      int
	uploaderName  string
	uploaderEmail string
	datasetID     string
	datasetFolder string
}

func NewConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configPath)
	v.ReadInConfig()

	cfg := &Config{
		emailAddress:  v.GetString("EMAIL_ADDRESS"),
		emailPassword: v.GetString("EMAIL_PASSWORD"),
		smtpHost:      v.GetString("SMTP_HOST"),
		smtpPort:      v.GetInt("SMTP_PORT"),
		uploaderName:  v.GetString("UPLOADER_NAME"),
		uploaderEmail: v.GetString("UPLOADER_EMAIL"),
		datasetID:     v.GetString("DATASET_ID"),
		datasetFolder: v.GetString("DATASET_FOLDER"),
	}
	return cfg, nil
}
