package optional_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reddec/dreaming-bard/internal/utils/optional"
)

func TestOptional_MarshalJSON(t *testing.T) {
	t.Run("WithValue", func(t *testing.T) {
		opt := optional.Optional[string]{Value: "hello", Set: true}
		data, err := json.Marshal(opt)
		require.NoError(t, err)
		assert.Equal(t, `"hello"`, string(data))
	})

	t.Run("WithoutValue", func(t *testing.T) {
		opt := optional.Optional[string]{Set: false}
		data, err := json.Marshal(opt)
		require.NoError(t, err)
		assert.Equal(t, "null", string(data))
	})

	t.Run("WithoutValueAndOmitted", func(t *testing.T) {
		type dataType struct {
			Data optional.Optional[string] `json:"data,omitzero"`
			OK   bool                      `json:"ok"`
		}

		data, err := json.Marshal(dataType{
			OK: true,
		})
		require.NoError(t, err)
		assert.Equal(t, `{"ok":true}`, string(data))
	})

	t.Run("WithIntValue", func(t *testing.T) {
		opt := optional.Optional[int]{Value: 123, Set: true}
		data, err := json.Marshal(opt)
		require.NoError(t, err)
		assert.Equal(t, `123`, string(data))
	})
}

func TestOptional_UnmarshalJSON(t *testing.T) {
	t.Run("WithValue", func(t *testing.T) {
		var opt optional.Optional[string]
		err := json.Unmarshal([]byte(`"hello"`), &opt)
		require.NoError(t, err)
		assert.True(t, opt.Set)
		assert.Equal(t, "hello", opt.Value)
	})

	t.Run("WithoutValue", func(t *testing.T) {
		var val struct {
			X   int                       `json:"x"`
			Opt optional.Optional[string] `json:"opt"`
		}
		err := json.Unmarshal([]byte(`{"x": 123}`), &val)
		require.NoError(t, err)
		assert.Equal(t, 123, val.X)
		assert.False(t, val.Opt.Set, "`opt` should not be set")
	})

	t.Run("WithIntValue", func(t *testing.T) {
		var opt optional.Optional[int]
		err := json.Unmarshal([]byte(`123`), &opt)
		require.NoError(t, err)
		assert.True(t, opt.Set)
		assert.Equal(t, 123, opt.Value)
	})
}
