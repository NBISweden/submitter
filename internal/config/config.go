package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	UserID        string `yaml:"UserID"`
	Uploader      string `yaml:"Uploader"`
	UploaderEmail string `yaml:"UploaderEmail"`
	DatasetID     string `yaml:"DatasetID"`
	DatasetFolder string `yaml:"DatasetFolder"`
	Email         string `yaml:"Email"`
	Password      string `yaml:"Password"`
	APIHost       string `yaml:"APIHost"`
	SMTPHost      string `yaml:"SMTPHost"`
	SMTPPort      int    `yaml:"SMTPPort"`
	S3Config      string `yaml:"S3Config"`
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
	return &c, nil
}

func (c *Config) GetAccessToken() (string, error) {
	file, err := os.Open(c.S3Config)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "access_token" {
			return value, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", nil
	}
	return "", fmt.Errorf("access_token not found in %s", c.S3Config)
}
