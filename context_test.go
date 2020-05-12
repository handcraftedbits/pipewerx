package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"bytes"
	"strings"

	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
)

//
// Testcases
//

// Context tests

var _ = g.Describe("Context", func() {
	g.Describe("given a new instance", func() {
		var buffer *bytes.Buffer
		var context Context

		g.BeforeEach(func() {
			buffer = new(bytes.Buffer)
		})

		g.Context("with no special configuration", func() {
			g.JustBeforeEach(func() {
				context = NewContext(ContextConfig{})

				context.Vars()["key"] = "value"
			})

			g.Describe("calling Copy", func() {
				g.It("should return a Context that is identical to the original one", func() {
					var copiedContext Context

					copiedContext = context.Copy()

					Expect(copiedContext.Log()).To(Equal(context.Log()))
					Expect(copiedContext.Vars()).To(HaveLen(len(context.Vars())))
					Expect(copiedContext.Vars()).To(HaveKey("key"))
					Expect(copiedContext.Vars()["key"]).To(Equal(context.Vars()["key"]))
				})
			})

			g.Describe("calling Vars", func() {
				g.It("should return the expected variables", func() {
					Expect(context.Vars()).To(HaveLen(1))
					Expect(context.Vars()).To(HaveKey("key"))
					Expect(context.Vars()["key"]).To(Equal("value"))
				})
			})
		})

		g.Context("with a log level specified", func() {
			g.Describe("calling Log", func() {
				g.JustBeforeEach(func() {
					context = NewContext(ContextConfig{
						Level:  zerolog.ErrorLevel,
						Writer: buffer,
					})
				})

				g.It("should respect the log level", func() {
					context.Log().Info().Msg("info")
					context.Log().Warn().Msg("warn")
					context.Log().Error().Msg("error")

					Expect(strings.TrimSpace(buffer.String())).To(HaveSuffix("error"))
				})

				g.It("should respect the log level after a new log level is set", func() {
					context.Log().Warn().Msg("warn")

					Expect(strings.TrimSpace(buffer.String())).To(BeEmpty())

					context.SetLogLevel(zerolog.WarnLevel)

					context.Log().Warn().Msg("warn")

					Expect(strings.TrimSpace(buffer.String())).To(HaveSuffix("warn"))
				})
			})
		})

		g.Context("with JSON output specified", func() {
			g.JustBeforeEach(func() {
				context = NewContext(ContextConfig{
					UseJSON: true,
					Writer:  buffer,
				})
			})

			g.Describe("calling Log", func() {
				g.It("should result in log statements being output in JSON", func() {
					context.Log().Warn().Msg("warn")

					Expect(buffer.String()).To(ContainSubstring("\"message\":\"warn\""))
				})
			})
		})
	})
})
