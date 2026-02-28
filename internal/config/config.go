package config

import "github.com/stickpro/kyp/pkg/logger"

type (
	Config struct {
		Log     logger.Config
		Storage StorageConfig
	}

	StorageConfig struct {
		DBPath string `yaml:"db_path" env:"KYP_DB_PATH"`
	}
)
