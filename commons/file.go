package commons

import (
	"io"
	"os"
	"path/filepath"
)

func CopyFile(src, dest string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer Close(source)

	err = os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		return err
	}
	destination, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer Close(destination)
	_, err = io.Copy(destination, source)
	return err
}