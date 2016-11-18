package execlib

import (
	"fmt"

	"github.com/zombiezen/mcm/catalog"
)

type Error struct {
	ResourceID      uint64
	ResourceComment string
	Err             error
	Output          []byte
}

func newError(e error) *Error {
	if e == nil {
		panic("newError should not be called with nil")
	}
	if e, ok := e.(*Error); ok {
		ee := new(Error)
		*ee = *e
		return ee
	}
	return &Error{Err: e}
}

func toError(e error) error {
	if e == nil {
		return nil
	}
	if _, ok := e.(*Error); ok {
		return e
	}
	return newError(e)
}

func errorf(format string, args ...interface{}) error {
	if len(args) > 0 {
		if e, ok := args[len(args)-1].(*Error); ok {
			ee := newError(e)
			args[len(args)-1] = e.Err
			ee.Err = fmt.Errorf(format, args...)
			return ee
		}
	}
	return newError(fmt.Errorf(format, args...))
}

func (e *Error) Error() string {
	if e.ResourceID == 0 {
		return e.Err.Error()
	}
	if e.ResourceComment == "" {
		return fmt.Sprintf("apply id=%d: %v", e.ResourceID, e.Err)
	}
	return fmt.Sprintf("apply %s (id=%d): %v", e.ResourceComment, e.ResourceID, e.Err)
}

func errorWithResource(r catalog.Resource, err error) error {
	if err == nil {
		return nil
	}
	e := newError(err)
	e.ResourceID = r.ID()
	e.ResourceComment, _ = r.Comment()
	return e
}

func errorWithOutput(out []byte, err error) error {
	if err == nil {
		return nil
	}
	e := newError(err)
	e.Output = out
	return e
}
