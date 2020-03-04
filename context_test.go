package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// Context tests

func TestContext_Copy(t *testing.T) {
	Convey("When copying a Context", t, func() {
		var context = newContext()

		Convey("the new Context's variables should be identical to the old Context", func() {
			var copiedContext Context

			context.Vars()["key"] = "value"

			copiedContext = context.Copy()

			So(copiedContext.Vars(), ShouldHaveLength, len(context.Vars()))
			So(copiedContext.Vars()["key"], ShouldEqual, context.Vars()["key"])
		})
	})
}
