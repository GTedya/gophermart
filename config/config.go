package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	RunAddress           string `mapstructure:"RUN_ADDRESS"`
	DatabaseURI          string `mapstructure:"DATABASE_URI"`
	AccrualSystemAddress string `mapstructure:"ACCRUAL_SYSTEM_ADDRESS"`
	Debug                bool   `mapstructure:"DEBUG"`
	SecretKey            string `mapstructure:"SECRET_KEY"`
}

func GetConfig() (c Config, err error) {
	pflag.StringVar(&c.RunAddress, "a", "localhost:8080", "address and port to run server")
	pflag.StringVar(&c.DatabaseURI, "d", "user=root password=root dbname=gophermart port=5432 sslmode=disable", "database dsn")
	pflag.StringVar(&c.AccrualSystemAddress, "r", "cmd/accrual/accrual_darwin_arm64", "accrual system address")
	pflag.StringVar(&c.SecretKey, "s", "some_key", "secret key")
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

	getFromEnvironment(&c)

	return c, nil
}

func getFromEnvironment(c *Config) {
	address, exists := os.LookupEnv("RUN_ADDRESS")
	if exists {
		c.RunAddress = address
	}
	db, exists := os.LookupEnv("DATABASE_URI")
	if exists {
		c.DatabaseURI = db
	}
	accrual, exists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if exists {
		c.AccrualSystemAddress = accrual
	}
}
