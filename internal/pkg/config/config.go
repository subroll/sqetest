package config

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	HTTPPort   = "http.port"
	DBAddress  = "db.address"
	DBName     = "db.name"
	DBUsername = "db.username"
	DBPassword = "db.password"

	fileName = "config"
	fileType = "json"
)

var configKeys = []string{HTTPPort, DBAddress, DBName, DBUsername, DBPassword}

func Load() error {
	viper.SetConfigName(fileName)
	viper.SetConfigType(fileType)
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err == nil {
		return err
	}

	if err := validate(); err != nil {
		return err
	}

	return nil
}

func validate() error {
	for _, configKey := range configKeys {
		configValue := viper.GetString(configKey)

		if configValue == "" {
			return fmt.Errorf("empty value for config key: %s", configKey)
		}
	}

	return nil
}
