package cfg

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	"github.com/cristalhq/aconfig"
)

func GenerateFlags(cfg any, opts ...Option) (string, error) {
	if reflect.ValueOf(cfg).Kind() != reflect.Ptr {
		return "", fmt.Errorf("config must be a pointer")
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
		return "", err
	}

	loader := aconfig.LoaderFor(cfg, aconf)

	if err := loader.Load(); err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	loader.Flags().SetOutput(buf)
	loader.Flags().PrintDefaults()

	return buf.String(), nil
}
