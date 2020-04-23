package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"
	"fmt"
)

//
// Public types
//

type MultiError interface {
	Causes() []error

	error
}

//
// Private types
//

// MultiError implementation.
type multiError struct {
	causes  []error
	message string
}

func (err *multiError) Causes() []error {
	return err.causes
}

func (err *multiError) Error() string {
	return err.message
}

//
// Private variables
//

var (
	errSourceNilFilesystem = errors.New("cannot create Source using nil Filesystem")
	errSourceNone          = errors.New("no Sources provided")
)

//
// Private functions
//

func newMultiError(message string, causes []error) MultiError {
	if causes == nil {
		causes = make([]error, 0)
	}

	return &multiError{
		causes:  causes,
		message: message,
	}
}

func newPanicError(value interface{}) error {
	var message = "a fatal error occurred: "

	switch val := value.(type) {
	case error:
		return fmt.Errorf(message+"%w", val)
	default:
		return fmt.Errorf(message+"%v", val)
	}
}
