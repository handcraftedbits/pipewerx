package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// MultiError tests

func TestMultiError(t *testing.T) {
	Convey("When creating a MultiError", t, func() {
		var err = newMultiError("message", []error{errors.New("error1"), errors.New("error2")})

		So(err, ShouldNotBeNil)

		Convey("calling Causes should return the expected value", func() {
			So(err.Causes(), ShouldNotBeNil)
			So(err.Causes(), ShouldHaveLength, 2)
			So(err.Causes()[0], ShouldNotBeNil)
			So(err.Causes()[0].Error(), ShouldEqual, "error1")
			So(err.Causes()[1], ShouldNotBeNil)
			So(err.Causes()[1].Error(), ShouldEqual, "error2")
		})

		Convey("calling Error should return the expected message", func() {
			So(err.Error(), ShouldEqual, "message")
		})
	})
}

func TestNewMultiError(t *testing.T) {
	Convey("When calling newMultiError with a nil causes array", t, func() {
		var err = newMultiError("message", nil)

		Convey("it should create a MultiError with an empty causes array", func() {
			So(err, ShouldNotBeNil)
			So(err.Causes(), ShouldNotBeNil)
			So(err.Causes(), ShouldHaveLength, 0)
		})
	})
}

// Other error test

func TestNewPanicError(t *testing.T) {
	Convey("When calling newPanicError with a non-error value", t, func() {
		var err = newPanicError("test")

		Convey("it should produce a simple error", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "a fatal error occurred: test")
		})
	})

	Convey("When calling newPanicError with an error value", t, func() {
		var err error
		var wrapped = errors.New("test")

		err = newPanicError(wrapped)

		Convey("it should produce an error with a wrapped value", func() {
			var unwrapped wrappedError

			So(err, ShouldNotBeNil)
			So(err, ShouldImplement, (*wrappedError)(nil))

			unwrapped = err.(wrappedError)

			So(unwrapped.Unwrap(), ShouldEqual, wrapped)
		})
	})
}

//
// Private types
//

type wrappedError interface {
	error

	Unwrap() error
}
