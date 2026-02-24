package cfg

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/cristalhq/aconfig"
)

const cellSeparator = "|"

func GenerateMarkdown(cfg any, filePath string, opts ...Option) error {
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

	return c.generateMarkdown(loader, filePath)
}

func (c *config) generateMarkdown(l *aconfig.Loader, filePath string) error {
	var table [][]string //nolint:prealloc

	table = append(table, []string{
		"**Name**", "**Required**", "**Secret**", "**Default value**", "**Usage**", "**Example**",
	})

	sizes := make([]int, len(table[0]))

	var lineSize int
	for i, cell := range table[0] {
		sizes[i] = utf8.RuneCountInString(cell) + 2
	}

	configFields := GetConfigFields(l)
	for _, f := range configFields {
		envName := f.EnvName
		if c.options.loaderConfig.EnvPrefix != "" {
			envName = c.options.loaderConfig.EnvPrefix + "_" + envName
		}

		cell := []string{
			"`" + envName + "`",
			boolIcon(f.IsRequired),
			boolIcon(f.IsSecret),
			codeBlock(f.DefaultValue),
			f.Usage,
			codeBlock(f.Example),
		}
		table = append(table, cell)

		lineSize = 0
		for i, item := range cell {
			if size := utf8.RuneCountInString(item); size+2 > sizes[i] {
				sizes[i] = size + 2
			}

			lineSize += sizes[i] // recalculate line size
		}
	}

	var out strings.Builder
	_, _ = out.WriteString("# Environments\n\n")
	for i, row := range table {
		_, _ = out.WriteString(cellSeparator)

		for j, cell := range row {
			size := utf8.RuneCountInString(" " + cell + " ")

			data := strings.Repeat(" ", sizes[j]-size)

			_, _ = out.WriteString(" " + cell + " ")
			_, _ = out.WriteString(data)

			if len(row)-1 != j {
				_, _ = out.WriteString(cellSeparator)
			}
		}

		if i == 0 {
			_, _ = out.WriteString(cellSeparator)
			_, _ = out.WriteRune('\n')

			_, _ = out.WriteString(cellSeparator)
			for j, item := range sizes {
				dashes := strings.Repeat("-", item)
				_, _ = out.WriteString(dashes)

				if len(sizes)-1 != j {
					_, _ = out.WriteString(cellSeparator)
				}
			}
		}

		_, _ = out.WriteString(cellSeparator)
		_, _ = out.WriteRune('\n')
	}

	_, _ = fmt.Fprintln(c.out, out.String())

	if filePath != "" {
		if err := os.WriteFile(filePath, []byte(out.String()), 0600); err != nil {
			return err
		}
	}

	return nil
}

func boolIcon(value bool) string {
	if value {
		return "✅"
	}

	return " "
}

func codeBlock(val string) string {
	if val == "" {
		return val
	}

	return "`" + val + "`"
}
