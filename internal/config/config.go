package config

import "github.com/kelseyhightower/envconfig"

// Config represents the application config.
// This makes it possible for the user to change the default path of the config files.
type Config struct {
	ConfigPath string `split_words:"true" default:"~/.hclvet.d"`
}

// FromEnv parses environment variables into the config object based on envconfig name
func FromEnv() (*Config, error) {
	var config Config

	err := envconfig.Process("hclvet", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
