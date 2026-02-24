package cfg

import (
	"fmt"
	"os"
	"strings"

	"github.com/cristalhq/aconfig"
)

func GenerateDefaultEnvs(cfg any, _ string, opts ...Option) error {
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

	loader.WalkFields(c.generateDefaultEnvs)

	return nil
}

func (c *config) generateDefaultEnvs(field aconfig.Field) bool {
	value := field.Tag("default")
	names := field.Tag("env")
	usage := field.Tag("usage")

	current := field
	if value == "" {
		value = "<empty>"
	}

	pad := 50

	var ok bool
	for {
		if current, ok = current.Parent(); !ok {
			break
		}

		names = fmt.Sprintf("%s_%s", current.Tag("env"), names)
	}

	var line strings.Builder
	_, _ = line.WriteString(names)
	_, _ = line.WriteString("=")
	_, _ = line.WriteString(value)

	if usage != "" {
		_, _ = line.WriteString(strings.Repeat(" ", pad-line.Len()))
		_, _ = line.WriteString("# " + usage)
	}

	c.print(line.String())

	return true
}
