package cfg

import (
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"slices"
	"strings"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigdotenv"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/stickpro/kyp/pkg/util/files"
)

var (
	boolTrueValues = []string{"true", "1"}
	fileDecoders   = map[string]aconfig.FileDecoder{
		".env":  aconfigdotenv.New(),
		".yaml": aconfigyaml.New(),
		".yml":  aconfigyaml.New(),
	}
)

type config struct {
	options *options

	out io.Writer

	args []string
	exit func(int)
}

// Load environment variables from `os env`, flags, `.env`, `.yaml` files and pass it to struct.
//
// For local development use `.env` file from root project.
//
// Load also call a `Validate` method if it proided.
//
// Example:
//
//	var config internalConfig.Config
//	if err := cfg.Load(&config); err != nil {
//		logger.Fatalf("could not load configuration: %v", err)
//	}
func Load(cfg any, opts ...Option) error {
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

func (c *config) print(value string) {
	_, _ = fmt.Fprintln(c.out, value)
}

func GetConfigFields(loader *aconfig.Loader) []ConfigField {
	res := []ConfigField{}

	loader.WalkFields(func(f aconfig.Field) bool {
		newField := ConfigField{
			Path:           f.Name(),
			DefaultValue:   f.Tag("default"),
			Usage:          f.Tag("usage"),
			Example:        f.Tag("example"),
			ValidateParams: f.Tag("validate"),
		}

		if slices.Contains(boolTrueValues, strings.ToLower(f.Tag("required"))) {
			newField.IsRequired = true
		}

		if slices.Contains(boolTrueValues, strings.ToLower(f.Tag("secret"))) {
			newField.IsSecret = true
		}

		if slices.Contains(boolTrueValues, strings.ToLower(f.Tag("disable_validation"))) {
			newField.DisableValidation = true
		}

		if strings.Contains(newField.ValidateParams, "required") {
			newField.IsRequired = true
		}

		envName := f.Tag("env")

		field := f
		var ok bool
		for {
			if field, ok = field.Parent(); !ok {
				break
			}

			envName = fmt.Sprintf("%s_%s", field.Tag("env"), envName)

			if !newField.DisableValidation &&
				slices.Contains(
					boolTrueValues,
					strings.ToLower(field.Tag("disable_validation")),
				) {
				newField.DisableValidation = true
			}
		}

		newField.EnvName = envName

		res = append(res, newField)

		return true
	})

	return res
}

// getAconfig return aconfig.Config based on options
func getAconfig(conf config) (aconfig.Config, error) {
	pwdDir, err := os.Getwd()
	if err != nil {
		return aconfig.Config{}, err
	}

	aconf := conf.options.loaderConfig
	aconf.FileDecoders = fileDecoders

	if conf.options.optionalFiles {
		existingFiles := make([]string, 0, len(aconf.Files))
		for _, file := range aconf.Files {
			absFilePath := file
			if !path.IsAbs(file) {
				absFilePath = path.Join(pwdDir, file)
			}
			if files.ExistsPath(absFilePath) {
				existingFiles = append(existingFiles, file)
			}
		}
		aconf.Files = existingFiles
	} else {
		for _, file := range aconf.Files {
			absFilePath := file
			if !path.IsAbs(file) {
				absFilePath = path.Join(pwdDir, file)
			}
			if !files.ExistsPath(absFilePath) {
				return aconfig.Config{}, fmt.Errorf("config file not found: %s", absFilePath)
			}
		}
	}

	return aconf, nil
}
