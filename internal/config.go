package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/goccy/go-yaml"
)

type Config struct {
	YandexDiskOAuthToken    string `env:"YANDEX_DISK_OAUTH_TOKEN,required"    yaml:"yandexDiskOauthToken"`
	YandexDiskProjectFolder string `env:"YANDEX_DISK_PROJECT_FOLDER,required" yaml:"yandexDiskProjectFolder"`
}

func LoadConfig() (*Config, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	configFilePath := filepath.Join(pwd, ".yadlfs.yaml")

	if _, err := os.Stat(configFilePath); err == nil {
		return loadConfigFromYAML(configFilePath)
	} else if os.IsNotExist(err) {
		return loadConfigFromEnv()
	}

	return nil, fmt.Errorf("failed to check for .yadlfs.yaml file: %w", err)
}

func loadConfigFromYAML(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML file: %w", err)
	}

	return &config, nil
}

func loadConfigFromEnv() (*Config, error) {
	var config Config
	opts := env.Options{RequiredIfNoDef: true} // Ensure required fields are set

	if err := env.ParseWithOptions(&config, opts); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &config, nil
}
