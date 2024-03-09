package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	RunAddress           string `mapstructure:"RUN_ADDRESS"`
	DatabaseUri          string `mapstructure:"DATABASE_URI"`
	AccrualSystemAddress string `mapstructure:"ACCRUAL_SYSTEM_ADDRESS"`
	Debug                bool   `mapstructure:"DEBUG"`
	SecretKey            string `mapstructure:"SECRET_KEY"`
}

func GetConfig() (c Config, err error) {
	pflag.StringVar(&c.RunAddress, "a", "localhost:8080", "address and port to run server")
	pflag.StringVar(&c.DatabaseUri, "d", "user=root password=root dbname=gophermart port=5432 sslmode=disable", "database dsn")
	pflag.StringVar(&c.AccrualSystemAddress, "r", "localhost:8080", "accrual system address")
	pflag.Parse()

	viper.SetConfigFile(".env")
	err = viper.ReadInConfig()
	if err != nil {
		return c, fmt.Errorf("viper reading error: %w", err)
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		return c, fmt.Errorf("viper unmarshalling error: %w", err)
	}

	return c, nil
}
