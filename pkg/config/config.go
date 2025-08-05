package config

import (
	"github.com/spf13/viper"
)

type GitAuthor struct {
	Name  string `mapstructure:"name"`
	Email string `mapstructure:"email"`
}

type GitConfig struct {
	Author GitAuthor `mapstructure:"author"`
}

type Config struct {
	Repositories        []string  `mapstructure:"repositories"`
	Git                 GitConfig `mapstructure:"git"`
	DeleteConflictedPRs bool      `mapstructure:"delete_conflicted_prs"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	var config Config

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Set default value for DeleteConflictedPRs if not specified
	if !viper.IsSet("delete_conflicted_prs") {
		config.DeleteConflictedPRs = true
	}

	return &config, nil
}
