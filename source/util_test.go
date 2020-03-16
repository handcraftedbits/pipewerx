package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"errors"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/client"
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

			stepper, err = newPathStepper(localFSInstance, "testdata/fileProducer/singleLevelSubdirs", true)

			So(err, ShouldBeNil)
			So(stepper, ShouldNotBeNil)

			stepper.fs = newErrorFilesystem(stepper.fs, nil, listFilesError)

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

				stepper, err = newPathStepper(newErrorFilesystem(nil, absolutePathError, nil), "/", false)

				So(stepper, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "absolutePath")
			})
	})
}

//
// Private types
//

// client.Filesystem implementation used to test uncommon errors.
type errorFilesystem struct {
	wrapped client.Filesystem

	absolutePathError error
	listFilesError    error
}

func (fs *errorFilesystem) AbsolutePath(path string) (string, error) {
	if fs.absolutePathError != nil {
		return "", fs.absolutePathError
	}

	return fs.wrapped.AbsolutePath(path)
}

func (fs *errorFilesystem) BasePart(path string) string {
	return fs.wrapped.BasePart(path)
}

func (fs *errorFilesystem) DirPart(path string) []string {
	return fs.wrapped.DirPart(path)
}

func (fs *errorFilesystem) ListFiles(path string) ([]os.FileInfo, error) {
	if fs.listFilesError != nil {
		return nil, fs.listFilesError
	}

	return fs.wrapped.ListFiles(path)
}

func (fs *errorFilesystem) PathSeparator() string {
	return fs.wrapped.PathSeparator()
}

func (fs *errorFilesystem) StatFile(path string) (os.FileInfo, error) {
	return fs.wrapped.StatFile(path)
}

//
// Private functions
//

func newErrorFilesystem(wrapped client.Filesystem, absolutePathError, listFilesError error) client.Filesystem {
	return &errorFilesystem{
		wrapped:           wrapped,
		absolutePathError: absolutePathError,
		listFilesError:    listFilesError,
	}
}
