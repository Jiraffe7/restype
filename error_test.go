package restype

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorsIs(t *testing.T) {
	{
		e := Error{
			kind: ErrorBodyFromRequest,
		}
		got := errors.Is(e, Error{})
		assert.True(t, got)
	}
	{
		e := Error{
			kind: ErrorExecuteRequest,
		}
		got := errors.Is(e, Error{})
		assert.True(t, got)
	}
	{
		e := Error{
			kind: ErrorResponseFromBytes,
		}
		got := errors.Is(e, Error{})
		assert.True(t, got)
	}
}

func TestErrorUnwrap(t *testing.T) {
	o := errors.New("original error")
	e := Error{original: o}
	got := errors.Unwrap(e)
	assert.Equal(t, o, got)
}
