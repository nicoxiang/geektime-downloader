package files

import (
	"errors"
	"os"
)

// CheckFileExists check if file exists
func CheckFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}
