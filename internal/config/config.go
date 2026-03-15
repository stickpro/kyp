package config

import (
	"time"

	"github.com/stickpro/kyp/pkg/logger"
)

type (
	Config struct {
		Log         logger.Config
		Storage     StorageConfig
		LockTimeout time.Duration `yaml:"lock_timeout" env:"KYP_LOCK_TIMEOUT" default:"5m"`
	}

	StorageConfig struct {
		DBPath string `yaml:"db_path" env:"KYP_DB_PATH"`
	}
)
