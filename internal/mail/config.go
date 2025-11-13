package mail

import (
	"github.com/spf13/viper"
)

type Config struct {
	EmailAddress  string
	EmailPassword string
	SMTPHost      string
	SMTPPort      int
	UploaderName  string
	UploaderEmail string
	DatasetID     string
	DatasetFolder string
}

func NewConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configPath)

	cfg := &Config{
		EmailAddress:  v.GetString("EMAIL_ADDRESS"),
		EmailPassword: v.GetString("EMAIL_PASSWORD"),
		SMTPHost:      v.GetString("SMTP_HOST"),
		SMTPPort:      v.GetInt("SMTP_PORT"),
		UploaderName:  v.GetString("UPLOADER_NAME"),
		UploaderEmail: v.GetString("UPLOADER_EMAIL"),
		DatasetID:     v.GetString("DATASET_ID"),
		DatasetFolder: v.GetString("DATASET_FOLDER"),
	}
	return cfg, nil
}
