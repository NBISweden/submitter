package config

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

type Config struct {
	DatasetFolder                string `mapstructure:"DATASET_FOLDER"`
	DatasetID                    string `mapstructure:"DATASET_ID"`
	UserID                       string `mapstructure:"USER_ID"`
	SslCaCert                    string `mapstructure:"SSL_CA_CERT"`
	Timeout                      int    `mapstructure:"JOB_TIMEOUT"`
	PollRate                     int    `mapstructure:"JOB_POLL_RATE"`
	JobDataDirectory             string `mapstructure:"JOB_DATA_DIRECTORY"`
	ClientAPIHost                string `mapstructure:"CLIENT_API_HOST"`
	ClientAccessToken            string `mapstructure:"CLIENT_ACCESS_TOKEN"`
	CertSecretName               string `mapstructure:"CERT_SECRET_NAME"`
	StorageSecretName            string `mapstructure:"STORAGE_SECRET_NAME"`
	MailAddress                  string `mapstructure:"MAIL_ADDRESS"`
	MailPassword                 string `mapstructure:"MAIL_PASSWORD"`
	MailSMTPHost                 string `mapstructure:"MAIL_SMTP_HOST"`
	MailSMTPPort                 int    `mapstructure:"MAIL_SMTP_PORT"`
	MailUploaderName             string `mapstructure:"MAIL_UPLOADER_NAME"`
	MailUploaderOrganizationName string `mapstructure:"MAIL_UPLOADER_ORGANIZATION_NAME"`
	MailUploader                 string `mapstructure:"MAIL_UPLOADER"`
	S3ArchiveEndpoint            string `mapstructure:"S3_ARCHIVE_ENDPOINT"`
	S3ArchiveBucket              string `mapstructure:"S3_ARCHIVE_BUCKET"`
	S3ArchiveAccessKey           string `mapstructure:"S3_ARCHIVE_ACCESS_KEY"`
	S3ArchiveSecretKey           string `mapstructure:"S3_ARCHIVE_SECRET_KEY"`
	S3MetadataEndpoint           string `mapstructure:"S3_METADATA_ENDPOINT"`
	S3MetadataBucket             string `mapstructure:"S3_METADATA_BUCKET"`
	S3MetadataAccessKey          string `mapstructure:"S3_METADATA_ACCESS_KEY"`
	S3MetadataSecretKey          string `mapstructure:"S3_METADATA_SECRET_KEY"`
}

func NewConfig(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(configPath)
	bindKeys(v)
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Info("config file not found, using environment only")
		} else {
			slog.Warn("failed to read config file, falling back to environment variables", "err", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func bindKeys(v *viper.Viper) {
	v.BindEnv("DATASET_FOLDER")
	v.BindEnv("DATASET_ID")
	v.BindEnv("USER_ID")
	v.BindEnv("SSL_CA_CERT")
	v.BindEnv("JOB_TIMEOUT")
	v.BindEnv("JOB_POLL_RATE")
	v.BindEnv("JOB_DATA_DIRECTORY")
	v.BindEnv("CLIENT_API_HOST")
	v.BindEnv("CLIENT_ACCESS_TOKEN")
	v.BindEnv("MAIL_ADDRESS")
	v.BindEnv("MAIL_PASSWORD")
	v.BindEnv("MAIL_SMTP_HOST")
	v.BindEnv("MAIL_SMTP_PORT")
	v.BindEnv("MAIL_UPLOADER_NAME")
	v.BindEnv("MAIL_UPLOADER_ORGANIZATION_NAME")
	v.BindEnv("MAIL_UPLOADER")
	v.BindEnv("CERT_SECRET_NAME")
	v.BindEnv("STORAGE_SECRET_NAME")
	v.BindEnv("S3_ARCHIVE_ENDPOINT")
	v.BindEnv("S3_ARCHIVE_BUCKET")
	v.BindEnv("S3_ARCHIVE_ACCESS_KEY")
	v.BindEnv("S3_ARCHIVE_SECRET_KEY")
	v.BindEnv("S3_METADATA_ENDPOINT")
	v.BindEnv("S3_METADATA_BUCKET")
	v.BindEnv("S3_METADATA_ACCESS_KEY")
	v.BindEnv("S3_METADATA_SECRET_KEY")
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("JOB_TIMEOUT", 4320)
	v.SetDefault("JOB_POLL_RATE", 180)
	v.SetDefault("JOB_DATA_DIRECTORY", "/data")

	v.SetDefault("CLIENT_API_HOST", "https://api.bp.nbis.se")
	v.SetDefault("CERT_SECRET_NAME", "sda-sda-svc-api-certs")
	v.SetDefault("STORAGE_SECRET_NAME", "sda-bpctl-storage")
	v.SetDefault("S3_METADATA_ENDPOINT", "storage.sto3.safedc.net")
	v.SetDefault("S3_METADATA_BUCKET", "public-metadata")
	v.SetDefault("S3_ARCHIVE_ENDPOINT", "s3a4.sto2.safedc.net")
	v.SetDefault("S3_ARCHIVE_BUCKET", "inbox-2024-01")
	v.SetDefault("MAIL_SMTP_HOST", "mail.nbis.se")
	v.SetDefault("MAIL_SMTP_PORT", 587)
}

func validateConfig(cfg *Config) error {
	if cfg.DatasetFolder == "" {
		return fmt.Errorf("DATASET_FOLDER requiered")
	}

	if cfg.DatasetID == "" {
		return fmt.Errorf("DATASET_ID requiered")
	}

	if cfg.UserID == "" {
		return fmt.Errorf("USER_ID requiered")
	}

	if cfg.PollRate > cfg.Timeout {
		return fmt.Errorf("JOB_POLL_RATE greater than JOB_TIMEOUT, set a pollrate that is less than the timeout value")
	}
	return nil
}
