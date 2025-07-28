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
	Repositories []string  `mapstructure:"repositories"`
	Git          GitConfig `mapstructure:"git"`
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

	return &config, nil
}
