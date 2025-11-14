package database

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	host         string
	port         int
	user         string
	password     string
	databaseName string
	schema       string
	caCert       string
	sslMode      string
	clientCert   string
	clientKey    string
}

func NewConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configPath)
	v.ReadInConfig()

	cfg := &Config{
		host:         v.GetString("DB_HOST"),
		port:         v.GetInt("DB_PORT"),
		user:         v.GetString("DB_USER"),
		password:     v.GetString("DB_PASSWORD"),
		databaseName: v.GetString("DB_NAME"),
		schema:       v.GetString("DB_SCHEMA"),
		caCert:       v.GetString("DB_CA_CERT"), // Can be shared in global config?
		sslMode:      v.GetString("DB_SSL_MODE"),
		clientCert:   v.GetString("DB_CLIENT_CERT"),
		clientKey:    v.GetString("DB_CLIENT_KEY"),
	}

	return cfg, nil
}

func (c *Config) dataSourceName() string {
	connInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.host, c.port, c.user, c.password, c.databaseName, c.sslMode)

	fmt.Println("connInfo: ", connInfo)

	if c.sslMode == "disable" {
		return connInfo
	}

	if c.caCert != "" {
		connInfo = fmt.Sprintf("%s sslrootcert=%s", connInfo, c.caCert)
	}

	if c.clientCert != "" {
		connInfo = fmt.Sprintf("%s sslcert=%s", connInfo, c.clientCert)
	}

	if c.clientKey != "" {
		connInfo = fmt.Sprintf("%s sslkey=%s", connInfo, c.clientKey)
	}

	return connInfo
}
