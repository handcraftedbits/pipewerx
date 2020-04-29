package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
			path: NewFilePath(nil, "name", "/"),
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
			path: NewFilePath(nil, "name.ext", "/"),
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

		Convey("calling Sys should return the expected value", func() {
			So(file1.Sys(), ShouldBeNil)
		})
	})
}

// FilePath tests

func TestFilePath(t *testing.T) {
	Convey("When creating a FilePath", t, func() {
		Convey("with a nil directory array", func() {
			var filePath = NewFilePath(nil, "name", "/")

			Convey("calling Dir should return a zero-length directory array", func() {
				So(filePath.Dir(), ShouldNotBeNil)
				So(filePath.Dir(), ShouldBeEmpty)
			})

			Convey("calling String should return a string that does not contain any separators", func() {
				So(filePath.String(), ShouldEqual, "name")
			})
		})

		Convey("with an empty directory array", func() {
			var filePath = NewFilePath([]string{}, "name", "/")

			Convey("calling Dir should return a zero-length directory array", func() {
				So(filePath.Dir(), ShouldNotBeNil)
				So(filePath.Dir(), ShouldBeEmpty)
			})

			Convey("calling String should return a string that does not contain any separators", func() {
				So(filePath.String(), ShouldEqual, "name")
			})
		})

		Convey("with a non-empty directory array", func() {
			var filePath = NewFilePath([]string{"home", "user"}, "name", "/")

			Convey("calling Dir should return the expected values", func() {
				So(filePath.Dir(), ShouldNotBeNil)
				So(filePath.Dir(), ShouldHaveLength, 2)
				So(filePath.Dir()[0], ShouldEqual, "home")
				So(filePath.Dir()[1], ShouldEqual, "user")
			})
		})

		Convey("for an extension-less file", func() {
			var filePath = NewFilePath(nil, "name", "/")

			Convey("calling Extension should return an empty string", func() {
				So(filePath.Extension(), ShouldBeEmpty)
				So(strings.Index(filePath.String(), "."), ShouldEqual, -1)
			})

			Convey("calling Name should return the correct filename", func() {
				So(filePath.Name(), ShouldEqual, "name")
			})
		})

		Convey("for a file with an extension", func() {
			var filePath = NewFilePath(nil, "name.ext", "/")

			Convey("calling Extension should return the correct extension", func() {
				So(filePath.Extension(), ShouldEqual, "ext")
				So(filePath.String(), ShouldEndWith, ".ext")
			})

			Convey("calling Name should return a filename that excludes the extension", func() {
				So(filePath.Name(), ShouldEqual, "name")
			})
		})

		Convey("for an absolute UNIX path", func() {
			var filePath = NewFilePath([]string{"/home", "user"}, "name.ext", "/")

			Convey("calling String should return the expected string representation", func() {
				So(filePath.String(), ShouldEqual, "/home/user/name.ext")
			})
		})

		Convey("for a relative UNIX path", func() {
			var filePath = NewFilePath([]string{"home", "user"}, "name.ext", "/")

			Convey("calling String should return the expected string representation", func() {
				So(filePath.String(), ShouldEqual, "home/user/name.ext")
			})
		})

		Convey("for an absolute Windows path", func() {
			var filePath = NewFilePath([]string{"C:", "home", "user"}, "name.ext", "\\")

			Convey("calling String should return the expected string representation", func() {
				So(filePath.String(), ShouldEqual, "C:\\home\\user\\name.ext")
			})
		})

		Convey("for a relative Windows path", func() {
			var filePath = NewFilePath([]string{"home", "user"}, "name.ext", "\\")

			Convey("calling String should return the expected string representation", func() {
				So(filePath.String(), ShouldEqual, "home\\user\\name.ext")
			})
		})
	})
}

//
// Private types
//

// In-memory Filesystem implementation.
type memFilesystem struct {
	absolutePathError error
	destroy           func() error
	listFilesError    error
	readFileError     error
	panic             bool
	root              *memFilesystemNode
	statFileError     error
}

func (fs *memFilesystem) AbsolutePath(path string) (string, error) {
	if fs.absolutePathError != nil {
		if fs.panic {
			panic(fs.absolutePathError)
		}

		return "", fs.absolutePathError
	}

	return path, nil
}

func (fs *memFilesystem) BasePart(path string) string {
	return filepath.Base(path)
}

func (fs *memFilesystem) Destroy() error {
	if fs.destroy != nil {
		return fs.destroy()
	}

	return nil
}

func (fs *memFilesystem) DirPart(path string) []string {
	var dirPart = filepath.Dir(path)

	if dirPart == "." {
		return []string{}
	}

	return strings.Split(dirPart, "/")
}

func (fs *memFilesystem) ListFiles(path string) ([]os.FileInfo, error) {
	var children []os.FileInfo
	var node *memFilesystemNode

	if fs.listFilesError != nil {
		if fs.panic {
			panic(fs.listFilesError)
		}

		return nil, fs.listFilesError
	}

	node = fs.findNode(path)

	if node == nil {
		return nil, memFilesystemErrorNotFound
	}

	if !node.IsDir() {
		return []os.FileInfo{node}, nil
	}

	for name, child := range node.children {
		// We don't want to redundantly provide node names when specifying them as literals, so we'll take this
		// opportunity to assign the correct name here.

		child.name = name

		children = append(children, child)
	}

	return children, nil
}

func (fs *memFilesystem) PathSeparator() string {
	return "/"
}

func (fs *memFilesystem) ReadFile(path string) (io.ReadCloser, error) {
	var node *memFilesystemNode
	var reader io.Reader

	if fs.readFileError != nil {
		if fs.panic {
			panic(fs.readFileError)
		}

		return nil, fs.readFileError
	}

	node = fs.findNode(path)

	if node == nil {
		return nil, memFilesystemErrorNotFound
	}

	reader = bytes.NewBufferString(node.contents)

	return ioutil.NopCloser(reader), nil
}

func (fs *memFilesystem) StatFile(path string) (os.FileInfo, error) {
	var node *memFilesystemNode

	if fs.statFileError != nil {
		if fs.panic {
			panic(fs.statFileError)
		}

		return nil, fs.statFileError
	}

	node = fs.findNode(path)

	if node == nil {
		return nil, memFilesystemErrorNotFound
	}

	return node, nil
}

func (fs *memFilesystem) findNode(path string) *memFilesystemNode {
	var curNode = fs.root
	var split []string

	if path == "" || path == "/" {
		return fs.root
	}

	split = strings.Split(path, "/")

	if split[0] == "" {
		split = split[1:]
	}

	for _, segment := range split {
		curNode = curNode.children[segment]

		if curNode == nil {
			return nil
		} else {
			// So we don't have to redundantly specify the name when we're creating memFilesystemNode literals in our
			// testcases.

			curNode.name = segment
		}
	}

	return curNode
}

// In-memory Filesystem node.
type memFilesystemNode struct {
	children map[string]*memFilesystemNode
	contents string
	name     string
}

func (node *memFilesystemNode) Name() string {
	return node.name
}

func (node *memFilesystemNode) Size() int64 {
	return 0
}

func (node *memFilesystemNode) Mode() os.FileMode {
	return os.ModePerm
}

func (node *memFilesystemNode) ModTime() time.Time {
	return time.Now()
}

func (node *memFilesystemNode) IsDir() bool {
	return len(node.children) > 0
}

func (node *memFilesystemNode) Sys() interface{} {
	return nil
}

//
// Private variables
//

var memFilesystemErrorNotFound = fmt.Errorf("path not found")
