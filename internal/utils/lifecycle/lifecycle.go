// Package lifecycle provides functionality for importing markdown documents from files and archives.
// It supports importing both individual markdown files and ZIP archives containing multiple markdown files.
// The package handles parsing markdown content into structured documents and delegates the actual import
// operation to a provided handler function.
package lifecycle

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/reddec/dreaming-bard/internal/utils/mark"
)

const MaxFormSize = 10 << 20

type ImportFunc[T any] func(document mark.Document[T], name string) error

func Import[T any](r *http.Request, handler ImportFunc[T]) error {
	if err := r.ParseMultipartForm(MaxFormSize); err != nil {
		return fmt.Errorf("parse multipart form: %w", err)
	}

	// supports both archives and individual files
	files := r.MultipartForm.File["file"]

	for _, file := range files {
		if isArchive(file) {
			if err := importArchive(file, handler); err != nil {
				return fmt.Errorf("import archive: %w", err)
			}
		} else if err := importFile(file, handler); err != nil {
			return fmt.Errorf("import file: %w", err)
		}
	}
	return nil
}

func importFile[T any](ref *multipart.FileHeader, handler ImportFunc[T]) error {
	f, err := ref.Open()
	if err != nil {
		return fmt.Errorf("open file %q: %w", ref.Filename, err)
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read file %q: %w", ref.Filename, err)
	}
	if err := importMarkdown(ref.Filename, content, handler); err != nil {
		return fmt.Errorf("import file %q: %w", ref.Filename, err)
	}
	return nil
}

func importArchive[T any](ref *multipart.FileHeader, handler ImportFunc[T]) error {
	f, err := ref.Open()
	if err != nil {
		return fmt.Errorf("open file %q: %w", ref.Filename, err)
	}
	defer f.Close()
	reader, err := zip.NewReader(f, ref.Size)
	if err != nil {
		return err
	}
	for _, file := range reader.File {
		stream, err := file.Open()
		if err != nil {
			return fmt.Errorf("open file %q in archive %q: %w", file.Name, ref.Filename, err)
		}
		content, err := io.ReadAll(stream)
		if err != nil {
			return fmt.Errorf("read file %q in archive %q: %w", file.Name, ref.Filename, err)
		}
		if err := importMarkdown(file.Name, content, handler); err != nil {
			return fmt.Errorf("import file %q in archive %q: %w", file.Name, ref.Filename, err)
		}
	}
	return nil
}

func importMarkdown[T any](name string, content []byte, handler ImportFunc[T]) error {
	d := mark.Parse[T](string(content))

	return handler(d, name)
}

func isArchive(h *multipart.FileHeader) bool {
	if strings.HasSuffix(h.Filename, ".zip") || h.Header.Get("Content-Type") == "application/zip" {
		return true
	}

	f, err := h.Open()
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = zip.NewReader(f, h.Size)
	return err == nil
}
