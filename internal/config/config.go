package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	UserID        string `yaml:"UserID"`
	Uploader      string `yaml:"Uploader"`
	DatasetID     string `yaml:"DatasetID"`
	DatasetFolder string `yaml:"DatasetFolder"`
	Email         string `yaml:"Email"`
	Password      string `yaml:"Password"`
	APIHost       string `yaml:"APIHost"`
	S3CmdConfig   string `yaml:"S3CmdConfig"`
	SMTPHost      string `yaml:"SMTPHost"`
	SMTPPort      int    `yaml:"SMTPPort"`
}

func NewConfig(configFilePath string) (*Config, error) {
	file, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var c Config
	err = yaml.Unmarshal(file, &c)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse configuration %v", err)
	}
	// We need underscores instead of @ signs in the email when calling SDA API
	c.enforceEmailFormat()
	return &c, nil
}

func (c *Config) enforceEmailFormat() {
	c.UserID = strings.ReplaceAll(c.UserID, "@", "_")
}
