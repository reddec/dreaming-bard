package dbo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// GetPef retrieves a preference by name, unmarshals its value into the specified type, and returns the result or an error.
func GetPef[T any](ctx context.Context, q *Queries, name string) (out T, err error) {
	v, err := q.GetPreference(ctx, name)
	if err != nil {
		return out, fmt.Errorf("get preference: %w", err)
	}
	return out, json.Unmarshal([]byte(v.Value), &out)
}

func SetPref[T any](ctx context.Context, q *Queries, name string, value T) error {
	v, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}
	return q.SetPreference(ctx, SetPreferenceParams{
		Name:  name,
		Value: string(v),
	})
}

// GetPrefWithDefault retrieves a preference value by name or returns the provided default if not found.
func GetPrefWithDefault[T any](ctx context.Context, q *Queries, name string, def T) (out T, err error) {
	v, err := GetPef[T](ctx, q, name)
	if err == nil {
		return v, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return def, nil
	}
	return v, err
}

func NewPref[T any](q *Queries, name string, def T) *Pref[T] {
	return &Pref[T]{
		name: name,
		q:    q,
		def:  def,
	}
}

type Pref[T any] struct {
	name string
	q    *Queries
	def  T
}

func (p *Pref[T]) Name() string {
	return p.name
}

func (p *Pref[T]) Get(ctx context.Context) (T, error) {
	return GetPrefWithDefault[T](ctx, p.q, p.name, p.def)
}

func (p *Pref[T]) Set(ctx context.Context, value T) error {
	return SetPref[T](ctx, p.q, p.name, value)
}
