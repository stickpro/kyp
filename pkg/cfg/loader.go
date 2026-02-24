package cfg

import (
	"fmt"
	"os"
	"reflect"

	"github.com/cristalhq/aconfig"
)

// GetConfigLoader returns aconfig loader instance
func GetConfigLoader(cfg any, opts ...Option) (*aconfig.Loader, error) {
	if reflect.ValueOf(cfg).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("config must be a pointer")
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
		return nil, err
	}

	loader := aconfig.LoaderFor(cfg, aconf)

	if err := loader.Load(); err != nil {
		return nil, err
	}

	return loader, nil
}
