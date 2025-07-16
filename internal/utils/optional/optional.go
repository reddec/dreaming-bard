package optional

import (
	"encoding/json"

	"github.com/goccy/go-yaml"
)

func WithDefault[T any](v T) Optional[T] {
	return Optional[T]{Value: v}
}

func With[T any](v T) Optional[T] {
	return Optional[T]{Value: v, Set: true}
}

type Optional[T any] struct {
	Value T
	Set   bool
}

func (opt *Optional[T]) UnmarshalYAML(bytes []byte) error {
	err := yaml.Unmarshal(bytes, &opt.Value)
	if err != nil {
		return err
	}
	opt.Set = true
	return nil
}

func (opt Optional[T]) MarshalYAML() ([]byte, error) {
	if !opt.Set {
		return nil, nil
	}
	return yaml.Marshal(opt.Value)
}

func (opt Optional[T]) MarshalJSON() ([]byte, error) {
	if !opt.Set {
		return []byte(""), nil
	}
	return json.Marshal(opt.Value)
}

func (opt *Optional[T]) UnmarshalJSON(bytes []byte) error {
	if bytes == nil {
		return nil
	}
	err := json.Unmarshal(bytes, &opt.Value)

	if err == nil {
		opt.Set = true
	}
	return err
}
