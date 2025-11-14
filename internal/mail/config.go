package mail

import (
	"github.com/NBISweden/submitter/internal/config"
	"github.com/spf13/viper"
)

type Config struct {
	emailAddress  string
	emailPassword string
	smtpHost      string
	smtpPort      int
	uploaderName  string
	uploaderMail  string
	datasetID     string
	datasetFolder string
}

func NewConfig(configPath string) (*Config, error) {
	globalConfig, _ := config.NewConfig(configPath)
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configPath)
	v.ReadInConfig()

	cfg := &Config{
		emailAddress:  v.GetString("MAIL_ADDRESS"),
		emailPassword: v.GetString("MAIL_PASSWORD"),
		smtpHost:      v.GetString("MAIL_SMTP_HOST"),
		smtpPort:      v.GetInt("MAIL_SMTP_PORT"),
		uploaderName:  v.GetString("MAIL_UPLOADER_NAME"),
		uploaderMail:  v.GetString("MAIL_UPLOADER"),
		datasetID:     globalConfig.DatasetID,
		datasetFolder: globalConfig.DatasetFolder,
	}
	return cfg, nil
}
