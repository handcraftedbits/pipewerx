package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"errors"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"golang.handcraftedbits.com/pipewerx"
)

//
// Testcases
//

// Edge case tests

func TestFindFiles(t *testing.T) {
	Convey("When calling findFiles", t, func() {
		Convey("it should return an error if the underlying filesystem throws an error when listing files", func() {
			var err error
			var listFilesError = errors.New("listFiles")
			var path pipewerx.FilePath
			var stepper *pathStepper

			stepper, err = newPathStepper(localFSInstance, "testdata/local/singleLevelSubdirs", true)

			So(err, ShouldBeNil)
			So(stepper, ShouldNotBeNil)

			stepper.fs = newErrorFilesystem(nil, listFilesError)

			path, err = stepper.nextFile()

			So(path, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "listFiles")
		})
	})
}

func TestStripRoot(t *testing.T) {
	Convey("When calling stripRoot with a path that does not start with the path separator", t, func() {
		var result = stripRoot("/abc", "/abcxyz", "/")

		Convey("it should return the path with only the root stripped from it", func() {
			So(result, ShouldEqual, "xyz")
		})
	})
}

func TestNewPathStepper(t *testing.T) {
	Convey("When calling newPathStepper", t, func() {
		Convey("it should return an error if the underlying filesystem throws an error when getting the absolute path",
			func() {
				var absolutePathError = errors.New("absolutePath")
				var err error
				var stepper *pathStepper

				stepper, err = newPathStepper(newErrorFilesystem(absolutePathError, nil), "/", false)

				So(stepper, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "absolutePath")
			})
	})
}

//
// Private types
//

// filesystem implementation used to test uncommon errors.
type errorFilesystem struct {
	*localFilesystem

	absolutePathError error
	listFilesError    error
}

func (fs *errorFilesystem) absolutePath(path string) (string, error) {
	if fs.absolutePathError != nil {
		return "", fs.absolutePathError
	}

	return fs.localFilesystem.absolutePath(path)
}

func (fs *errorFilesystem) listFiles(path string) ([]os.FileInfo, error) {
	if fs.listFilesError != nil {
		return nil, fs.listFilesError
	}

	return fs.localFilesystem.listFiles(path)
}

//
// Private functions
//

func newErrorFilesystem(absolutePathError, listFilesError error) filesystem {
	return &errorFilesystem{
		localFilesystem:   localFSInstance,
		absolutePathError: absolutePathError,
		listFilesError:    listFilesError,
	}
}
