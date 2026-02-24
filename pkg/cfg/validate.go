package cfg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/cristalhq/aconfig"
	"github.com/go-playground/validator/v10"
	"github.com/stickpro/go-store/pkg/util/structs"
)

// ValidateConfig validates config struct with environment variables and custom validation functions.
func ValidateConfig(cfg any, opts ...Option) error {
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

	if err := c.validateEnvs(cfg, loader); err != nil {
		return err
	}

	c.print("OK")

	return nil
}

func (c *config) validateEnvs(cfg any, loader *aconfig.Loader) error {
	if err := validateElem(c.options.ctx, cfg); err != nil {
		return err
	}

	val := reflect.ValueOf(cfg).Elem()
	for i := 0; i < val.NumField(); i++ {
		if err := validateElem(c.options.ctx, val.Field(i).Addr().Interface()); err != nil {
			return err
		}
	}

	// init validator
	validate := validator.New()
	for _, item := range c.options.validateFuncs {
		_ = validate.RegisterValidation(item.Tag, item.Fn, item.CallValidationEvenIfNull...)
	}

	// validate struct
	errs := []string{}
	configFields := GetConfigFields(loader)
	for _, f := range configFields {
		if f.DisableValidation || f.ValidateParams == "" {
			continue
		}

		fieldValue, err := structs.LookupString(cfg, f.Path)
		if err != nil {
			return err
		}

		if err := validate.Var(fieldValue.Interface(), f.ValidateParams); err != nil {
			errs = append(errs,
				strings.ReplaceAll(
					err.Error(),
					"Key: '' Error:Field validation for ''",
					fmt.Sprintf("Validate %s env error:", f.EnvName),
				),
			)
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

func validateElem(ctx context.Context, elem any) error {
	// try to validate with Validate() error
	if tmp, ok := elem.(interface {
		Validate() error
	}); ok {
		if err := tmp.Validate(); err != nil {
			return err
		}
	}

	// try to validate with Validate(ctx context.Context) error
	if ctx != nil {
		if tmp, ok := elem.(interface {
			Validate(ctx context.Context) error
		}); ok {
			if err := tmp.Validate(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}
