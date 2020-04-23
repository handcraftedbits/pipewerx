package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// Context tests

func TestContext(t *testing.T) {
	Convey("When creating a Context", t, func() {
		var context = NewContext(ContextConfig{})

		context.Vars()["key"] = "value"

		Convey("with a log level specified", func() {
			var buffer bytes.Buffer

			context = NewContext(ContextConfig{
				Level:  zerolog.ErrorLevel,
				Writer: &buffer,
			})

			Convey("the logger should abide by the log level", func() {
				context.Log().Info().Msg("info")
				context.Log().Warn().Msg("warn")
				context.Log().Error().Msg("error")

				So(strings.TrimSpace(buffer.String()), ShouldEndWith, "error")
			})

			Convey("after setting a new log level the logger should abide by the new log level", func() {
				context.Log().Warn().Msg("warn")

				So(strings.TrimSpace(buffer.String()), ShouldBeEmpty)

				context.SetLogLevel(zerolog.WarnLevel)

				context.Log().Warn().Msg("warn")

				So(strings.TrimSpace(buffer.String()), ShouldEndWith, "warn")
			})
		})

		Convey("and specifying that JSON output should be used", func() {
			var buffer bytes.Buffer

			context = NewContext(ContextConfig{
				UseJSON: true,
				Writer:  &buffer,
			})

			Convey("the log output should be in JSON", func() {
				context.Log().Warn().Msg("warn")

				So(buffer.String(), ShouldContainSubstring, "\"message\":\"warn\"")
			})
		})

		Convey("calling Copy should return a Context that is identical to the old Context", func() {
			var copiedContext Context

			copiedContext = context.Copy()

			So(copiedContext.Log(), ShouldEqual, context.Log())
			So(copiedContext.Vars(), ShouldHaveLength, len(context.Vars()))
			So(copiedContext.Vars()["key"], ShouldEqual, context.Vars()["key"])
		})

		Convey("calling Vars should return the expected value", func() {
			So(context.Vars(), ShouldHaveLength, 1)
			So(context.Vars()["key"], ShouldEqual, "value")
		})
	})
}
