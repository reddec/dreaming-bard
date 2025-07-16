package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type Type int

const (
	TypeObject Type = iota
	TypeArray
	TypeBoolean
	TypeInteger
	TypeNumber
	TypeString
)

type Function struct {
	name         string
	description  string
	inputSchema  *Schema
	outputSchema *Schema
	readOnly     bool
	handler      func(ctx context.Context, values map[string]any) (json.RawMessage, error)
}

func (f *Function) Name() string {
	return f.name
}

func (f *Function) Description() string {
	return f.description
}

func (f *Function) IsReadOnly() bool {
	return f.readOnly
}

func (f *Function) CallJSON(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var values map[string]any
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}
	return f.handler(ctx, values)
}

func Tool[T any, Out any](handler func(ctx context.Context, args T) (Out, error)) *FunctionBuilder {
	return &FunctionBuilder{
		fn: &Function{
			inputSchema:  Scan(reflect.TypeFor[T]()),
			outputSchema: Scan(reflect.TypeFor[Out]()),
			handler: func(ctx context.Context, values map[string]any) (json.RawMessage, error) {
				var input T
				decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
					Metadata: nil,
					Result:   &input,
					TagName:  "json",
				})
				if err != nil {
					return nil, fmt.Errorf("create decoder: %w", err)
				}
				if err := decoder.Decode(values); err != nil {
					return nil, err
				}

				out, err := handler(ctx, input)
				if err != nil {
					return nil, fmt.Errorf("call tool: %w", err)
				}
				return json.Marshal(out)
			},
		},
	}
}

type FunctionBuilder struct {
	fn *Function
}

func (b *FunctionBuilder) Name(name string) *FunctionBuilder {
	b.fn.name = name
	return b
}

func (b *FunctionBuilder) Description(description string) *FunctionBuilder {
	b.fn.description = description
	return b
}

func (b *FunctionBuilder) ReadOnly() *FunctionBuilder {
	b.fn.readOnly = true
	return b
}

func (b *FunctionBuilder) Build() (*Function, error) {
	if b.fn.inputSchema.Type != TypeObject {
		return nil, fmt.Errorf("input type must be struct")
	}
	cp := *b.fn
	return &cp, nil
}

func Scan(value reflect.Type) *Schema {
	if value == nil {
		return &Schema{Type: TypeObject}
	}

	return scanReflectType(value)
}

type Schema struct {
	Type        Type
	Fields      map[string]*Schema // for type == TypeObject only
	Items       *Schema            // for type == TypeArray only
	Description string             // mostly for fields
}

func (s *Schema) ToOpenAPI() map[string]any {
	result := make(map[string]any)

	switch s.Type {
	case TypeObject:
		result["type"] = "object"
		if len(s.Fields) == 0 {
			break
		}
		properties := make(map[string]any, len(s.Fields))
		required := make([]string, 0, len(s.Fields))

		for fieldName, fieldSchema := range s.Fields {
			properties[fieldName] = fieldSchema.ToOpenAPI()
			required = append(required, fieldName)
		}

		result["properties"] = properties
		result["required"] = required
	case TypeArray:
		result["type"] = "array"
		if s.Items != nil {
			result["items"] = s.Items.ToOpenAPI()
		}
	case TypeBoolean:
		result["type"] = "boolean"
	case TypeInteger:
		result["type"] = "integer"
		result["format"] = "int64"
	case TypeNumber:
		result["type"] = "number"
	case TypeString:
		result["type"] = "string"
	}

	if s.Description != "" {
		result["description"] = s.Description
	}

	return result
}

func scanReflectType(t reflect.Type) *Schema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Bool:
		return &Schema{Type: TypeBoolean}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: TypeInteger}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: TypeNumber}
	case reflect.String:
		return &Schema{Type: TypeString}
	case reflect.Slice, reflect.Array:
		elemType := t.Elem()
		return &Schema{
			Type:  TypeArray,
			Items: scanReflectType(elemType),
		}
	case reflect.Struct:
		fields := make(map[string]*Schema)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			fieldName, skip := getJSONFieldName(field)
			if skip {
				continue
			}

			schema := scanReflectType(field.Type)
			if desc := field.Tag.Get("description"); desc != "" {
				schema.Description = desc
			}

			fields[fieldName] = schema

		}
		return &Schema{
			Type:   TypeObject,
			Fields: fields,
		}
	case reflect.Map:
		// treat map as object :-(
		fallthrough
	default:
		return &Schema{Type: TypeObject}
	}
}

func getJSONFieldName(field reflect.StructField) (name string, skip bool) {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return field.Name, false
	}

	if jsonTag == "-" {
		return "", true
	}

	parts := strings.Split(jsonTag, ",")
	name = parts[0]

	if name == "" {
		return field.Name, false
	}

	return name, false
}
