package xfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func AtomicWrite(targetPath string, handler func(out io.Writer) error) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0700); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(targetPath), filepath.Base(targetPath)+".tmp.*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := handler(tmpFile); err != nil {
		return errors.Join(err, tmpFile.Close(), os.RemoveAll(tmpFile.Name()))
	}

	if err := tmpFile.Close(); err != nil {
		return errors.Join(fmt.Errorf("close temporary file: %w", err), os.Remove(tmpFile.Name()))
	}

	if err := os.Rename(tmpFile.Name(), targetPath); err != nil {

		return errors.Join(fmt.Errorf("rename temporary file: %w", err), os.Remove(tmpFile.Name()))
	}

	return nil
}

func ValidateName(name string) error {
	if strings.ContainsAny(name, "\u0000/.\\") {
		return fmt.Errorf("name can not contain any of the following characters: (dot), (NULL), (forward slash), (backward slash)")
	}
	return nil
}
