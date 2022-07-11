package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

func init() {
	viper.MustBindEnv("postgres.host", "POSTGRES_HOST")
	viper.MustBindEnv("postgres.port", "POSTGRES_PORT")
	viper.MustBindEnv("postgres.user", "POSTGRES_USER")
	viper.MustBindEnv("postgres.password", "POSTGRES_PASSWORD")
	viper.MustBindEnv("postgres.dbname", "POSTGRES_DBNAME")
}

func (config PostgresConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
	)
}
