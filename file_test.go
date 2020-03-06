package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"fmt"
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

// FileEvaluator implementation that discards all files
type discardingFileEvaluator struct {
}

func (evaluator *discardingFileEvaluator) Destroy() error {
	return nil
}

func (evaluator *discardingFileEvaluator) ShouldKeep(file File) (bool, error) {
	return false, nil
}

// Test File implementation
type testFile struct {
	path FilePath
}

func (file *testFile) Path() FilePath {
	return file.path
}

func (file *testFile) Reader() (io.ReadCloser, error) {
	return nil, nil
}

// Test FileEvaluator implementation
type testFileEvaluator struct {
	destroyError   error
	index          int
	keepError      error
	maxEvaluations int
}

func (evaluator *testFileEvaluator) Destroy() error {
	return evaluator.destroyError
}

func (evaluator *testFileEvaluator) ShouldKeep(file File) (bool, error) {
	evaluator.index++

	if evaluator.index == evaluator.maxEvaluations {
		return false, evaluator.keepError
	}

	return true, nil
}

// Test FileProducer implementation
type testFileProducer struct {
	destroyError error
	files        []File
	index        int
	nextError    error
}

func (producer *testFileProducer) Destroy() error {
	if producer.destroyError != nil {
		return producer.destroyError
	}

	return nil
}

func (producer *testFileProducer) Next() (File, error) {
	producer.index++

	if producer.index == len(producer.files) {
		if producer.nextError != nil {
			return nil, producer.nextError
		}

		return nil, nil
	}

	if producer.index > len(producer.files) {
		return nil, nil
	}

	return producer.files[producer.index], nil
}

//
// Private functions
//

func newSimpleFileProducer(prefix string, size int) FileProducer {
	return newTestFileProducer(prefix, size, nil, nil)
}

func newTestFileEvaluator(maxEvaluations int, destroyError, keepError error) FileEvaluator {
	return &testFileEvaluator{
		destroyError:   destroyError,
		index:          -1,
		keepError:      keepError,
		maxEvaluations: maxEvaluations,
	}
}

func newTestFileProducer(prefix string, size int, destroyError, nextError error) FileProducer {
	var files = make([]File, size)

	for i := 0; i < size; i++ {
		files[i] = &testFile{
			path: newTestFilePath(fmt.Sprintf("%s%d", prefix, i)),
		}
	}

	return &testFileProducer{
		destroyError: destroyError,
		files:        files,
		index:        -1,
		nextError:    nextError,
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
