package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"

	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Testcases
//

// MultiError tests

var _ = g.Describe("MultiError", func() {
	g.Describe("calling newMultiError", func() {
		var err MultiError

		g.Context("with a nil causes array", func() {
			g.BeforeEach(func() {
				err = newMultiError("message", nil)
			})

			g.It("should succeed", func() {
				Expect(err).NotTo(BeNil())
			})

			g.Describe("and then calling Causes", func() {
				g.It("should return an empty causes array", func() {
					Expect(err.Causes()).NotTo(BeNil())
					Expect(err.Causes()).To(HaveLen(0))
				})
			})
		})

		g.Context("with a normal causes array", func() {
			g.BeforeEach(func() {
				err = newMultiError("message", []error{errors.New("error1"), errors.New("error2")})
			})

			g.It("should succeed", func() {
				Expect(err).NotTo(BeNil())
			})

			g.Describe("and then calling Causes", func() {
				g.It("should return the expected set of errors", func() {
					Expect(err.Causes()).NotTo(BeNil())
					Expect(err.Causes()).To(HaveLen(2))
					Expect(err.Causes()[0]).NotTo(BeNil())
					Expect(err.Causes()[0].Error()).To(Equal("error1"))
					Expect(err.Causes()[1]).NotTo(BeNil())
					Expect(err.Causes()[1].Error()).To(Equal("error2"))
				})
			})

			g.Describe("calling Error", func() {
				g.It("should return the expected error message", func() {
					Expect(err.Error()).To(Equal("message"))
				})
			})
		})
	})
})

// Other error tests

var _ = g.Describe("newPanicError", func() {
	g.Describe("calling newPanicError", func() {
		var err error

		g.Context("with a non-error value", func() {
			g.BeforeEach(func() {
				err = newPanicError("test")
			})

			g.It("should return a normal error", func() {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("a fatal error occurred: test"))
			})
		})

		g.Context("with an error value", func() {
			var wrapped error

			g.BeforeEach(func() {
				wrapped = errors.New("test")
				err = newPanicError(wrapped)
			})

			g.It("should return an error with a wrapped value", func() {
				var ok bool
				var unwrapped wrappedError

				Expect(err).NotTo(BeNil())

				unwrapped, ok = err.(wrappedError)

				Expect(ok).To(BeTrue())
				Expect(unwrapped.Unwrap()).To(Equal(wrapped))
			})
		})
	})
})

//
// Private types
//

// Simple interface used to unwrap a wrapped error returned from fmt.Errorf().
type wrappedError interface {
	error

	Unwrap() error
}
