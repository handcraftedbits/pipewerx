package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"io"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// FilePath tests

func TestFilePath_Dir(t *testing.T) {
	Convey("When creating a FilePath with a nil directory array", t, func() {
		var filePath = NewFilePath(nil, "name", "/")

		Convey("it should have a zero-length directory array", func() {
			So(filePath.Dir(), ShouldNotBeNil)
			So(filePath.Dir(), ShouldBeEmpty)
		})

		Convey("its string representation should not contain any separators", func() {
			So(filePath.String(), ShouldEqual, "name")
		})
	})

	Convey("When creating a FilePath with a non-nil but empty directory array", t, func() {
		var filePath = NewFilePath([]string{}, "name", "/")

		Convey("its string representation should not contain any separators", func() {
			So(filePath.String(), ShouldEqual, "name")
		})
	})

	Convey("When creating a FilePath with a non-nil, non-empty directory array", t, func() {
		var filePath = NewFilePath([]string{"home", "user"}, "name", "/")

		Convey("the expected values should appear in the directory array", func() {
			So(filePath.Dir(), ShouldNotBeNil)
			So(filePath.Dir(), ShouldHaveLength, 2)
			So(filePath.Dir()[0], ShouldEqual, "home")
			So(filePath.Dir()[1], ShouldEqual, "user")
		})
	})
}

func TestFilePath_Extension(t *testing.T) {
	Convey("When creating a FilePath with an extension-less file", t, func() {
		var filePath = NewFilePath(nil, "name", "/")

		Convey("it should have an empty extension value", func() {
			So(filePath.Extension(), ShouldBeEmpty)
			So(strings.Index(filePath.String(), "."), ShouldEqual, -1)
		})

		Convey("it should have the correct filename", func() {
			So(filePath.Name(), ShouldEqual, "name")
		})
	})

	Convey("When creating a FilePath for a file with an extension", t, func() {
		var filePath = NewFilePath(nil, "name.ext", "/")

		Convey("it should have the correct extension", func() {
			So(filePath.Extension(), ShouldEqual, "ext")
			So(filePath.String(), ShouldEndWith, ".ext")
		})

		Convey("it should have a filename that excludes the extension", func() {
			So(filePath.Name(), ShouldEqual, "name")
		})
	})
}

func TestFilePath_Name(t *testing.T) {
	Convey("When creating a FilePath with an extension-less file", t, func() {
		var filePath = NewFilePath(nil, "name", "/")

		Convey("it should have the correct filename", func() {
			So(filePath.Name(), ShouldEqual, "name")
		})
	})

	Convey("When creating a FilePath for a file with an extension", t, func() {
		var filePath = NewFilePath(nil, "name.ext", "/")

		Convey("it should have a filename that excludes the extension", func() {
			So(filePath.Name(), ShouldEqual, "name")
		})
	})

}

func TestFilePath_String(t *testing.T) {
	Convey("When dealing with UNIX paths", t, func() {
		Convey("When creating a valid, absolute FilePath", func() {
			var filePath = NewFilePath([]string{"/home", "user"}, "name.ext", "/")

			Convey("its string representation should be as expected", func() {
				So(filePath.String(), ShouldEqual, "/home/user/name.ext")
			})
		})

		Convey("When creating a valid, relative FilePath", func() {
			var filePath = NewFilePath([]string{"home", "user"}, "name.ext", "/")

			Convey("its string representation should be as expected", func() {
				So(filePath.String(), ShouldEqual, "home/user/name.ext")
			})
		})
	})

	Convey("When dealing with Windows paths", t, func() {
		Convey("When creating a valid, absolute FilePath", func() {
			var filePath = NewFilePath([]string{"C:", "home", "user"}, "name.ext", "\\")

			Convey("its string representation should be as expected", func() {
				So(filePath.String(), ShouldEqual, "C:\\home\\user\\name.ext")
			})
		})

		Convey("When creating a valid, relative FilePath", func() {
			var filePath = NewFilePath([]string{"home", "user"}, "name.ext", "\\")

			Convey("its string representation should be as expected", func() {
				So(filePath.String(), ShouldEqual, "home\\user\\name.ext")
			})
		})
	})
}

//
// Private types
//

// File implementation used for testing
type testFile struct {
	path FilePath
}

func (file *testFile) Delete() error {
	return nil
}

func (file *testFile) Path() FilePath {
	return file.path
}

func (file *testFile) Reader() (io.ReadCloser, error) {
	return nil, nil
}

func (file *testFile) Rename(newPath FilePath) error {
	return nil
}

func (file *testFile) Writer() (io.WriteCloser, error) {
	return nil, nil
}

//
// Private functions
//

func newTestFile(path string) File {
	return &testFile{
		path: newTestFilePath(path),
	}
}

func newTestFilePath(path string) FilePath {
	var split = strings.Split(path, "/")

	if split[0] == "" {
		split = split[1:]
		split[0] = "/" + split[0]
	}

	return NewFilePath(split[:len(split)-1], split[len(split)-1], "/")
}
