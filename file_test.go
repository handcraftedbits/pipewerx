package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// File tests

func TestFile(t *testing.T) {
	Convey("When creating a File", t, func() {
		var file1 = &file{
			fileInfo: &nilFileInfo{
				name: "name",
			},
			fs: &memFilesystem{
				root: &memFilesystemNode{
					children: map[string]*memFilesystemNode{
						"name": {
							contents: "abc",
						},
					},
				},
			},
			path: newFilePath(nil, "name", "/"),
		}
		var file2 = &file{
			fileInfo: &nilFileInfo{
				name: "name.ext",
			},
			fs: &memFilesystem{
				root: &memFilesystemNode{
					children: map[string]*memFilesystemNode{
						"name.ext": {
							contents: "abc",
						},
					},
				},
			},
			path: newFilePath(nil, "name.ext", "/"),
		}

		Convey("calling IsDir should return the expected value", func() {
			So(file1.IsDir(), ShouldBeFalse)
		})

		Convey("calling Mode should return the expected value", func() {
			So(file1.Mode(), ShouldEqual, os.ModePerm)
		})

		Convey("calling ModTime should return the expected value", func() {
			So(file1.ModTime(), ShouldNotBeNil)
		})

		Convey("calling Name", func() {
			Convey("for an a File with no extension should return the expected value", func() {
				So(file1.Name(), ShouldEqual, "name")
			})

			Convey("for an a File with an extension should return the expected value", func() {
				So(file2.Name(), ShouldEqual, "name.ext")
			})
		})

		Convey("calling Path should return the expected FilePath", func() {
			So(file1.Path().String(), ShouldEqual, "name")
		})

		Convey("calling ReadFile should return the expected contents", func() {
			var contents []byte
			var err error
			var reader io.ReadCloser

			reader, err = file1.Reader()

			So(err, ShouldBeNil)
			So(reader, ShouldNotBeNil)

			contents, err = ioutil.ReadAll(reader)

			So(err, ShouldBeNil)
			So(string(contents), ShouldEqual, "abc")
		})

		// TODO: use something more descriptive than "expected value" where appropriate.
		Convey("calling Size should return the expected value", func() {
			So(file1.Size(), ShouldEqual, 0)
		})

		Convey("calling Sys should return the expected value", func() {
			So(file1.Sys(), ShouldBeNil)
		})
	})
}

// FilePath tests

func TestFilePath(t *testing.T) {
	Convey("When creating a FilePath", t, func() {
		var filePath FilePath

		Convey("with a nil directory array", func() {
			filePath = newFilePath(nil, "name", "/")

			Convey("calling Dir should return a zero-length directory array", func() {
				So(filePath.Dir(), ShouldNotBeNil)
				So(filePath.Dir(), ShouldBeEmpty)
			})

			Convey("calling String should return a string that does not contain any separators", func() {
				So(filePath.String(), ShouldEqual, "name")
			})
		})

		Convey("with an empty directory array", func() {
			filePath = newFilePath([]string{}, "name", "/")

			Convey("calling Dir should return a zero-length directory array", func() {
				So(filePath.Dir(), ShouldNotBeNil)
				So(filePath.Dir(), ShouldBeEmpty)
			})

			Convey("calling String should return a string that does not contain any separators", func() {
				So(filePath.String(), ShouldEqual, "name")
			})
		})

		Convey("with a non-empty directory array", func() {
			filePath = newFilePath([]string{"home", "user"}, "name", "/")

			Convey("calling Dir should return the expected values", func() {
				So(filePath.Dir(), ShouldNotBeNil)
				So(filePath.Dir(), ShouldHaveLength, 2)
				So(filePath.Dir()[0], ShouldEqual, "home")
				So(filePath.Dir()[1], ShouldEqual, "user")
			})
		})

		Convey("for an extension-less file", func() {
			filePath = newFilePath(nil, "name", "/")

			Convey("calling Extension should return an empty string", func() {
				So(filePath.Extension(), ShouldBeEmpty)
				So(strings.Index(filePath.String(), "."), ShouldEqual, -1)
			})

			Convey("calling Name should return the correct filename", func() {
				So(filePath.Name(), ShouldEqual, "name")
			})
		})

		Convey("for a file with an extension", func() {
			filePath = newFilePath(nil, "name.ext", "/")

			Convey("calling Extension should return the correct extension", func() {
				So(filePath.Extension(), ShouldEqual, "ext")
				So(filePath.String(), ShouldEndWith, ".ext")
			})

			Convey("calling Name should return a filename that excludes the extension", func() {
				So(filePath.Name(), ShouldEqual, "name")
			})
		})

		Convey("for an absolute UNIX path", func() {
			filePath = newFilePath([]string{"/home", "user"}, "name.ext", "/")

			Convey("calling String should return the expected string representation", func() {
				So(filePath.String(), ShouldEqual, "/home/user/name.ext")
			})
		})

		Convey("for a relative UNIX path", func() {
			filePath = newFilePath([]string{"home", "user"}, "name.ext", "/")

			Convey("calling String should return the expected string representation", func() {
				So(filePath.String(), ShouldEqual, "home/user/name.ext")
			})
		})

		Convey("for an absolute Windows path", func() {
			filePath = newFilePath([]string{"C:", "home", "user"}, "name.ext", "\\")

			Convey("calling String should return the expected string representation", func() {
				So(filePath.String(), ShouldEqual, "C:\\home\\user\\name.ext")
			})
		})

		Convey("for a relative Windows path", func() {
			filePath = newFilePath([]string{"home", "user"}, "name.ext", "\\")

			Convey("calling String should return the expected string representation", func() {
				So(filePath.String(), ShouldEqual, "home\\user\\name.ext")
			})
		})
	})
}
