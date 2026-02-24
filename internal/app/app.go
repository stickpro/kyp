package app

import (
	"context"

	"github.com/stickpro/kyp/internal/config"
	"github.com/stickpro/kyp/pkg/logger"
)

func Run(_ context.Context, _ *config.Config, l logger.Logger) {
	l.Info("Hello world")
}
