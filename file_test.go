package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Testcases
//

// File tests

var _ = g.Describe("File", func() {
	g.Describe("given a new instance", func() {
		g.Context("for a file with a Filesystem that returns an error when calling ReadFile", func() {
			var f File

			g.BeforeEach(func() {
				f = &file{
					fileInfo: &nilFileInfo{
						name: "name",
						size: 3,
					},
					fs: &memFilesystem{
						readFileError: errors.New("readFile"),
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
			})

			g.Describe("calling Reader", func() {
				g.It("should return an error", func() {
					var err error
					var reader io.ReadCloser

					reader, err = f.Reader()

					Expect(err).NotTo(BeNil())
					Expect(reader).To(BeNil())
				})
			})
		})

		g.Context("for a file with no extension", func() {
			var f *file

			g.BeforeEach(func() {
				f = &file{
					fileInfo: &nilFileInfo{
						name: "name",
						size: 3,
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
			})

			g.Describe("calling IsDir", func() {
				g.It("should return false", func() {
					Expect(f.IsDir()).To(BeFalse())
				})
			})

			g.Describe("calling Mode", func() {
				g.It("should return the expected mode", func() {
					Expect(f.Mode()).To(Equal(os.ModePerm))
				})
			})

			g.Describe("calling ModTime", func() {
				g.It("should not return nil", func() {
					Expect(f.ModTime()).NotTo(BeNil())
				})
			})

			g.Describe("calling Name", func() {
				g.It("should return the expected filename", func() {
					Expect(f.Name()).To(Equal("name"))
				})
			})

			g.Describe("calling Path", func() {
				g.It("should return the expected path", func() {
					Expect(f.Path().String()).To(Equal("name"))
				})
			})

			g.Describe("calling Reader", func() {
				var sink *testEventSink

				g.JustBeforeEach(func() {
					sink = newTestEventSink()

					RegisterEventSink(sink)

					f.sourceID = sink.id
				})

				g.It("should return the expected file contents and send the appropriate events", func() {
					var contents []byte
					var err error
					var reader io.ReadCloser

					reader, err = f.Reader()

					Expect(reader).NotTo(BeNil())
					Expect(err).To(BeNil())

					contents, err = ioutil.ReadAll(reader)

					Expect(string(contents)).To(Equal("abc"))
					Expect(err).To(BeNil())

					Expect(reader.Close()).To(BeNil())

					// Have to check these events manually since we're checking for specific data within them.

					Expect(sink.events).To(HaveLen(3))
					Expect(sink.events[0].Component()).To(Equal(componentFile))
					Expect(sink.events[0].Data()).NotTo(BeNil())
					Expect(sink.events[0].Data()).To(HaveKey(eventFieldLength))
					Expect(sink.events[0].Data()[eventFieldLength]).To(BeEquivalentTo(3))
					Expect(sink.events[0].Type()).To(Equal(eventTypeOpened))
					Expect(sink.events[1].Component()).To(Equal(componentFile))
					Expect(sink.events[1].Data()).NotTo(BeNil())
					Expect(sink.events[1].Data()).To(HaveKey(eventFieldLength))
					Expect(sink.events[1].Data()[eventFieldLength]).To(BeEquivalentTo(3))
					Expect(sink.events[1].Type()).To(Equal(eventTypeRead))
					Expect(sink.events[2].Component()).To(Equal(componentFile))
					Expect(sink.events[2].Data()).NotTo(BeNil())
					Expect(sink.events[2].Type()).To(Equal(eventTypeClosed))
				})
			})

			g.Describe("calling Size", func() {
				g.It("should return the expected size", func() {
					Expect(f.Size()).To(BeEquivalentTo(3))
				})
			})

			g.Describe("calling Sys", func() {
				g.It("should return nil", func() {
					Expect(f.Sys()).To(BeNil())
				})
			})
		})

		g.Context("for a file with an extension", func() {
			var f File

			g.BeforeEach(func() {
				f = &file{
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
			})

			g.Describe("calling Name", func() {
				g.It("should return the expected filename", func() {
					Expect(f.Name()).To(Equal("name.ext"))
				})
			})
		})
	})
})

var _ = g.Describe("File events", func() {
	var f = &file{
		fileInfo: &nilFileInfo{
			name: "name.ext",
			size: 10,
		},
		path: newFilePath(nil, "name.ext", "/"),
	}
	var sourceID = "source"

	g.Describe("calling fileEventClosed", func() {
		g.It("should return a valid event", func() {
			Expect(fileEventClosed(f, sourceID)).To(beAValidFileEvent(eventTypeClosed, sourceID, f.Path().String(), 0))
		})
	})

	g.Describe("calling fileEventOpened", func() {
		g.It("should return a valid event", func() {
			Expect(fileEventOpened(f, sourceID)).To(beAValidFileEvent(eventTypeOpened, sourceID, f.Path().String(), 10))
		})
	})

	g.Describe("calling fileEventRead", func() {
		g.It("should return a valid event", func() {
			Expect(fileEventRead(f, sourceID, 5)).To(beAValidFileEvent(eventTypeRead, sourceID, f.Path().String(), 5))
		})
	})
})

// FilePath tests

var _ = g.Describe("FilePath", func() {
	g.Describe("given a new instance", func() {
		var filePath FilePath

		g.Context("created with a nil directory array", func() {
			g.BeforeEach(func() {
				filePath = newFilePath(nil, "name", "/")
			})

			g.Describe("calling Dir", func() {
				g.It("should return a zero-length directory array", func() {
					Expect(filePath.Dir()).NotTo(BeNil())
					Expect(filePath.Dir()).To(BeEmpty())
				})
			})

			g.Describe("calling String", func() {
				g.It("should return a string that does not contain any separators", func() {
					Expect(filePath.String()).To(Equal("name"))
				})
			})
		})

		g.Context("created with an empty directory array", func() {
			g.BeforeEach(func() {
				filePath = newFilePath([]string{}, "name", "/")
			})

			g.Describe("calling Dir", func() {
				g.It("should return a zero-length directory array", func() {
					Expect(filePath.Dir()).NotTo(BeNil())
					Expect(filePath.Dir()).To(BeEmpty())
				})
			})

			g.Describe("calling String", func() {
				g.It("should return a string that does not contain any separators", func() {
					Expect(filePath.String()).To(Equal("name"))
				})
			})
		})

		g.Context("created with a non-empty directory array", func() {
			g.BeforeEach(func() {
				filePath = newFilePath([]string{"home", "user"}, "name", "/")
			})

			g.Describe("calling Dir", func() {
				g.It("should return the expected directory values", func() {
					Expect(filePath.Dir()).NotTo(BeNil())
					Expect(filePath.Dir()).To(HaveLen(2))
					Expect(filePath.Dir()[0]).To(Equal("home"))
					Expect(filePath.Dir()[1]).To(Equal("user"))
				})
			})
		})

		g.Context("created with a file that does not have an extension", func() {
			g.BeforeEach(func() {
				filePath = newFilePath(nil, "name", "/")
			})

			g.Describe("calling Extension", func() {
				g.It("should return an empty string", func() {
					Expect(filePath.Extension()).To(BeEmpty())
					Expect(strings.Index(filePath.String(), ".")).To(Equal(-1))
				})
			})

			g.Describe("calling Name", func() {
				g.It("should return the correct filename", func() {
					Expect(filePath.Name()).To(Equal("name"))
				})
			})
		})

		g.Context("created with a file that has an extension", func() {
			g.BeforeEach(func() {
				filePath = newFilePath(nil, "name.ext", "/")
			})

			g.Describe("calling Extension", func() {
				g.It("should return the correct extension", func() {
					Expect(filePath.Extension()).To(Equal("ext"))
					Expect(filePath.String()).To(HaveSuffix(".ext"))
				})
			})

			g.Describe("calling Name", func() {
				g.It("should return a filename that excludes the extension", func() {
					Expect(filePath.Name()).To(Equal("name"))
				})
			})
		})

		g.Context("created with an absolute UNIX path", func() {
			g.BeforeEach(func() {
				filePath = newFilePath([]string{"/home", "user"}, "name.ext", "/")
			})

			g.Describe("calling String", func() {
				g.It("should return the expected string representation", func() {
					Expect(filePath.String()).To(Equal("/home/user/name.ext"))
				})
			})
		})

		g.Context("created with a relative UNIX path", func() {
			g.BeforeEach(func() {
				filePath = newFilePath([]string{"home", "user"}, "name.ext", "/")
			})

			g.Describe("calling String", func() {
				g.It("should return the expected string representation", func() {
					Expect(filePath.String()).To(Equal("home/user/name.ext"))
				})
			})
		})

		g.Context("created with an absolute Windows path", func() {
			g.BeforeEach(func() {
				filePath = newFilePath([]string{"C:", "home", "user"}, "name.ext", "\\")
			})

			g.Describe("calling String", func() {
				g.It("should return the expected string representation", func() {
					Expect(filePath.String()).To(Equal("C:\\home\\user\\name.ext"))
				})
			})
		})

		g.Context("created with a relative Windows path", func() {
			g.BeforeEach(func() {
				filePath = newFilePath([]string{"home", "user"}, "name.ext", "\\")
			})

			g.Describe("calling String", func() {
				g.It("should return the expected string representation", func() {
					Expect(filePath.String()).To(Equal("home\\user\\name.ext"))
				})
			})
		})
	})
})

//
// Private types
//

// FileEvaluator implementation that discards files based on file extension.
type extensionFileEvaluator struct {
	destroyError    error
	extension       string
	panic           bool
	shouldKeepError error
}

func (evaluator *extensionFileEvaluator) Destroy() error {
	if evaluator.panic {
		panic(evaluator.destroyError)
	}

	return evaluator.destroyError
}

func (evaluator *extensionFileEvaluator) ShouldKeep(file File) (bool, error) {
	if evaluator.shouldKeepError != nil {
		if evaluator.panic {
			panic(evaluator.shouldKeepError)
		}

		return false, evaluator.shouldKeepError
	}

	if file.Path().Extension() == evaluator.extension {
		return true, nil
	}

	return false, nil
}

// In-memory Filesystem implementation.
type memFilesystem struct {
	FilesystemDefaults

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

	return fs.FilesystemDefaults.AbsolutePath(path)
}

func (fs *memFilesystem) Destroy() error {
	if fs.destroy != nil {
		return fs.destroy()
	}

	return nil
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

// Dummy os.FileInfo implementation used to test file and fileInfoStack.
type nilFileInfo struct {
	name string
	size int64
}

func (fi *nilFileInfo) Name() string {
	return fi.name
}

func (fi *nilFileInfo) Size() int64 {
	return fi.size
}

func (fi *nilFileInfo) Mode() os.FileMode {
	return os.ModePerm
}

func (fi *nilFileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi *nilFileInfo) IsDir() bool {
	return false
}

func (fi *nilFileInfo) Sys() interface{} {
	return nil
}

//
// Private variables
//

var memFilesystemErrorNotFound = fmt.Errorf("path not found")
