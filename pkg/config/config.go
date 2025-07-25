package config

import (
	"github.com/spf13/viper"
)

type Repository struct {
	Name string `mapstructure:"name"`
	URL  string `mapstructure:"url"`
}

type Config struct {
	Repositories []Repository `mapstructure:"repositories"`
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
