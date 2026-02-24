package cfg

import (
	"fmt"
	"os"
	"reflect"

	"github.com/cristalhq/aconfig"
)

// LoadForTests environment variables from `os env`, flags, `.env`, `.yaml` files and pass it to struct.
//
// Disabled flags detection and any cli features.
//
// For local development use `.env` file from root project.
//
// LoadForTests also call a `Validate` method if it proided.
//
// Example:
//
//	var config internalConfig.Config
//	if err := cfg.LoadForTests(&config); err != nil {
//		logger.Fatalf("could not load configuration: %v", err)
//	}
func LoadForTests(cfg any, opts ...Option) error {
	if reflect.ValueOf(cfg).Kind() != reflect.Ptr {
		return fmt.Errorf("config must be a pointer")
	}

	options := newOptions(opts...)

	c := config{
		out:     os.Stdout,
		exit:    os.Exit,
		args:    os.Args[1:],
		options: options,
	}

	aconf, err := getAconfig(c)
	if err != nil {
		return err
	}

	loader := aconfig.LoaderFor(cfg, aconf)

	if err := loader.Load(); err != nil {
		return err
	}

	if options.validate {
		if err := c.validateEnvs(cfg, loader); err != nil {
			return err
		}
	}

	return nil
}
