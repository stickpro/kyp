package files

import (
	"errors"
	"os"
)

func ExistsPath(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}
