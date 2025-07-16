package views

import (
	"sync"
)

type ErrorParams struct {
	Error error
}

var viewError = sync.OnceValue(func() *DynamicView[ErrorParams] {
	return Inherit[ErrorParams](getBase(), "error.gohtml")
})

func Error(data ErrorParams) *Builder[ErrorParams] {
	return viewError().OK(data)
}
