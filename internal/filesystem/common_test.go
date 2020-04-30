package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// fileInfo tests

// smbFileInfo tests

func TestFileInfo(t *testing.T) {
	var now = time.Now()

	Convey("When creating a fileInfo", t, func() {
		var fi = &fileInfo{
			mode:    os.ModeDir,
			modTime: now,
			name:    "name",
			size:    1,
		}

		Convey("calling IsDir should return the expected value", func() {
			So(fi.IsDir(), ShouldBeTrue)
		})

		Convey("calling Mode should return the expected value", func() {
			So(fi.Mode(), ShouldEqual, os.ModeDir)
		})

		Convey("calling ModTime should return the expected value", func() {
			So(fi.ModTime(), ShouldEqual, now)
		})

		Convey("calling Name should return the expected value", func() {
			So(fi.Name(), ShouldEqual, "name")
		})

		Convey("calling Size should return the expected value", func() {
			So(fi.Size(), ShouldEqual, 1)
		})

		Convey("calling Sys should return the expected value", func() {
			So(fi.Sys(), ShouldEqual, nil)
		})
	})
}
