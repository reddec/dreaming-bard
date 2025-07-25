// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package common

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

const (
	// PurposeWrite is a Purpose of type write.
	PurposeWrite Purpose = "write"
	// PurposeSummary is a Purpose of type summary.
	PurposeSummary Purpose = "summary"
	// PurposeEnhance is a Purpose of type enhance.
	PurposeEnhance Purpose = "enhance"
	// PurposePlan is a Purpose of type plan.
	PurposePlan Purpose = "plan"
)

var ErrInvalidPurpose = errors.New("not a valid Purpose")

// PurposeValues returns a list of the values for Purpose
func PurposeValues() []Purpose {
	return []Purpose{
		PurposeWrite,
		PurposeSummary,
		PurposeEnhance,
		PurposePlan,
	}
}

// String implements the Stringer interface.
func (x Purpose) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x Purpose) IsValid() bool {
	_, err := ParsePurpose(string(x))
	return err == nil
}

var _PurposeValue = map[string]Purpose{
	"write":   PurposeWrite,
	"summary": PurposeSummary,
	"enhance": PurposeEnhance,
	"plan":    PurposePlan,
}

// ParsePurpose attempts to convert a string to a Purpose.
func ParsePurpose(name string) (Purpose, error) {
	if x, ok := _PurposeValue[name]; ok {
		return x, nil
	}
	return Purpose(""), fmt.Errorf("%s is %w", name, ErrInvalidPurpose)
}

// MarshalText implements the text marshaller method.
func (x Purpose) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *Purpose) UnmarshalText(text []byte) error {
	tmp, err := ParsePurpose(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

var errPurposeNilPtr = errors.New("value pointer is nil") // one per type for package clashes

// Scan implements the Scanner interface.
func (x *Purpose) Scan(value interface{}) (err error) {
	if value == nil {
		*x = Purpose("")
		return
	}

	// A wider range of scannable types.
	// driver.Value values at the top of the list for expediency
	switch v := value.(type) {
	case string:
		*x, err = ParsePurpose(v)
	case []byte:
		*x, err = ParsePurpose(string(v))
	case Purpose:
		*x = v
	case *Purpose:
		if v == nil {
			return errPurposeNilPtr
		}
		*x = *v
	case *string:
		if v == nil {
			return errPurposeNilPtr
		}
		*x, err = ParsePurpose(*v)
	default:
		return errors.New("invalid type for Purpose")
	}

	return
}

// Value implements the driver Valuer interface.
func (x Purpose) Value() (driver.Value, error) {
	return x.String(), nil
}

const (
	// RoleUser is a Role of type user.
	RoleUser Role = "user"
	// RoleAssistant is a Role of type assistant.
	RoleAssistant Role = "assistant"
	// RoleToolCall is a Role of type tool_call.
	RoleToolCall Role = "tool_call"
	// RoleToolResult is a Role of type tool_result.
	RoleToolResult Role = "tool_result"
)

var ErrInvalidRole = errors.New("not a valid Role")

// RoleValues returns a list of the values for Role
func RoleValues() []Role {
	return []Role{
		RoleUser,
		RoleAssistant,
		RoleToolCall,
		RoleToolResult,
	}
}

// String implements the Stringer interface.
func (x Role) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x Role) IsValid() bool {
	_, err := ParseRole(string(x))
	return err == nil
}

var _RoleValue = map[string]Role{
	"user":        RoleUser,
	"assistant":   RoleAssistant,
	"tool_call":   RoleToolCall,
	"tool_result": RoleToolResult,
}

// ParseRole attempts to convert a string to a Role.
func ParseRole(name string) (Role, error) {
	if x, ok := _RoleValue[name]; ok {
		return x, nil
	}
	return Role(""), fmt.Errorf("%s is %w", name, ErrInvalidRole)
}

// MarshalText implements the text marshaller method.
func (x Role) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *Role) UnmarshalText(text []byte) error {
	tmp, err := ParseRole(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

var errRoleNilPtr = errors.New("value pointer is nil") // one per type for package clashes

// Scan implements the Scanner interface.
func (x *Role) Scan(value interface{}) (err error) {
	if value == nil {
		*x = Role("")
		return
	}

	// A wider range of scannable types.
	// driver.Value values at the top of the list for expediency
	switch v := value.(type) {
	case string:
		*x, err = ParseRole(v)
	case []byte:
		*x, err = ParseRole(string(v))
	case Role:
		*x = v
	case *Role:
		if v == nil {
			return errRoleNilPtr
		}
		*x = *v
	case *string:
		if v == nil {
			return errRoleNilPtr
		}
		*x, err = ParseRole(*v)
	default:
		return errors.New("invalid type for Role")
	}

	return
}

// Value implements the driver Valuer interface.
func (x Role) Value() (driver.Value, error) {
	return x.String(), nil
}
