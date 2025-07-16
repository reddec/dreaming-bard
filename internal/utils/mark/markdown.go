package mark

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const delimiter = "---\n"

func Build[T any](metadata T, content string) (string, error) {
	d := &Document[T]{
		Metadata: metadata,
		Content:  content,
	}
	return d.Build()
}

type Document[T any] struct {
	Metadata T
	Content  string
}

func (d *Document[T]) Build() (string, error) {
	data, err := yaml.Marshal(d.Metadata)
	if err != nil {
		return "", err
	}
	return delimiter + string(data) + delimiter + "\n" + d.Content, nil
}

func Parse[T any](content string) Document[T] {
	// TODO: implement streaming parser

	if !strings.HasPrefix(content, delimiter) {
		return Document[T]{
			Content: content,
		}
	}

	idx := strings.Index(content, "\n"+delimiter)
	if idx == -1 {
		// no second delimiter
		return Document[T]{
			Content: content,
		}
	}

	header := strings.TrimSpace(content[len(delimiter):idx])
	body := strings.TrimSpace(content[idx+len(delimiter):])

	var meta T
	if err := yaml.Unmarshal([]byte(header), &meta); err != nil {
		// broken metadata
		return Document[T]{
			Content: content,
		}
	}

	return Document[T]{
		Metadata: meta,
		Content:  body,
	}
}

func ParseFile[T any](path string) (Document[T], error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Document[T]{}, err
	}
	return Parse[T](string(content)), nil
}

type NamedDocument[T any] struct {
	Name     string
	Category string
	Document[T]
}

func LoadDirectory[T any](root string) ([]NamedDocument[T], error) {
	var docs []NamedDocument[T]
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		doc, err := ParseFile[T](path)
		if err != nil {
			return err
		}
		name := cleanExtension(filepath.Base(path))

		category, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}

		docs = append(docs, NamedDocument[T]{
			Name:     name,
			Category: category,
			Document: doc,
		})
		return nil
	})
	return docs, err
}

func cleanExtension(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}
