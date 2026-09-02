package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port           string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	JWTSecret      string
	JWTExpiry      time.Duration
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOUseSSL    bool
	MinIOBucket    string
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	viper.SetDefault("PORT", "3000")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_EXPIRY_MINUTES", 30)
	viper.SetDefault("MINIO_USE_SSL", false)
	viper.SetDefault("MINIO_BUCKET", "supplyhub")

	return &Config{
		Port:           viper.GetString("PORT"),
		DBHost:         viper.GetString("DB_HOST"),
		DBPort:         viper.GetString("DB_PORT"),
		DBUser:         viper.GetString("DB_USER"),
		DBPassword:     viper.GetString("DB_PASSWORD"),
		DBName:         viper.GetString("DB_NAME"),
		DBSSLMode:      viper.GetString("DB_SSLMODE"),
		JWTSecret:      viper.GetString("JWT_SECRET"),
		JWTExpiry:      time.Duration(viper.GetInt("JWT_EXPIRY_MINUTES")) * time.Minute,
		MinIOEndpoint:  viper.GetString("MINIO_ENDPOINT"),
		MinIOAccessKey: viper.GetString("MINIO_ACCESS_KEY"),
		MinIOSecretKey: viper.GetString("MINIO_SECRET_KEY"),
		MinIOUseSSL:    viper.GetBool("MINIO_USE_SSL"),
		MinIOBucket:    viper.GetString("MINIO_BUCKET"),
	}
}
