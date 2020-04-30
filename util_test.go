package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"bytes"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// cancellationHelper tests

func TestCancellationHelper(t *testing.T) {
	Convey("When creating a cancellationHelper", t, func() {
		var buffer bytes.Buffer
		var context = NewContext(ContextConfig{Writer: &buffer})
		var helper = newCancellationHelper(context.Log(), nil, nil, nil)

		Convey("calling finalize", func() {
			Convey("should log a warning if a panic occurs", func() {
				helper.finalize()

				So(buffer.String(), ShouldContainSubstring, "an unexpected error occurred during cancellation")
			})

			Convey("should call the cancellation callback even if a panic occurs", func() {
				var x = 1

				helper.callback = func() {
					x = 2
				}

				helper.finalize()

				So(buffer.String(), ShouldContainSubstring, "an unexpected error occurred during cancellation")
				So(x, ShouldEqual, 2)
			})

			Convey("should log a warning if a panic occurs in the invoker's callback function", func() {
				helper = newCancellationHelper(context.Log(), make(chan Result), make(chan<- struct{}), nil)

				helper.callback = func() {
					panic("invoker panic")
				}

				helper.finalize()

				So(buffer.String(), ShouldContainSubstring, "invoker panic")
			})
		})
	})
}

// pathStepper tests

func TestNewPathStepper(t *testing.T) {
	Convey("When calling newPathStepper", t, func() {
		var err error
		var stepper *pathStepper

		Convey("it should return an error if the underlying filesystem throws an error when getting the absolute path",
			func() {
				stepper, err = newPathStepper(&memFilesystem{
					absolutePathError: errors.New("absolutePath"),
				}, "/", false)

				So(stepper, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "absolutePath")
			})
	})
}

func TestPathStepper(t *testing.T) {
	Convey("When creating a pathStepper", t, func() {
		var err error
		var stepper *pathStepper

		Convey("for a single file", func() {
			stepper, err = newPathStepper(&memFilesystem{
				root: &memFilesystemNode{
					children: map[string]*memFilesystemNode{
						"file": {},
					},
				},
			}, "/file", false)

			So(err, ShouldBeNil)
			So(stepper, ShouldNotBeNil)

			Convey("calling nextFile should return a single file", func() {
				var file File

				So(err, ShouldBeNil)
				So(stepper, ShouldNotBeNil)

				file, err = stepper.nextFile()

				So(err, ShouldBeNil)
				So(file, ShouldNotBeNil)
				So(file.Path().String(), ShouldEqual, "file")

				file, err = stepper.nextFile()

				So(err, ShouldBeNil)
				So(file, ShouldBeNil)
			})
		})

		Convey("for a nested directory structure", func() {
			stepper, err = newPathStepper(&memFilesystem{
				root: &memFilesystemNode{
					children: map[string]*memFilesystemNode{
						"dir1": {
							children: map[string]*memFilesystemNode{
								"file1": {},
							},
						},
						"dir2": {
							children: map[string]*memFilesystemNode{
								"file2": {},
							},
						},
						"file3": {},
					},
				},
			}, "/", true)

			So(err, ShouldBeNil)
			So(stepper, ShouldNotBeNil)

			Convey("calling nextFile should return the expected files", func() {
				var file File

				for i := 0; i < 3; i++ {
					file, err = stepper.nextFile()

					So(err, ShouldBeNil)
					So(file, ShouldNotBeNil)
				}

				// And make sure calling it again will give us no results.

				file, err = stepper.nextFile()

				So(err, ShouldBeNil)
				So(file, ShouldBeNil)
			})
		})
	})
}

// stepperFileStack tests

func TestStepperFileStack(t *testing.T) {
	Convey("When creating a stepperFileStack", t, func() {
		var stack = &stepperFileStack{}

		Convey("calling clear should remove all items", func() {
			stack.push(&stepperFile{
				fileInfo: &nilFileInfo{name: "a"},
				path:     "a",
			})
			stack.push(&stepperFile{
				fileInfo: &nilFileInfo{name: "b"},
				path:     "b",
			})

			So(stack.isEmpty(), ShouldBeFalse)

			stack.clear()

			So(stack.isEmpty(), ShouldBeTrue)
		})

		Convey("calling peek should return the last item on the stack", func() {
			stack.push(&stepperFile{
				fileInfo: &nilFileInfo{name: "a"},
				path:     "a",
			})
			stack.push(&stepperFile{
				fileInfo: &nilFileInfo{name: "b"},
				path:     "b",
			})

			So(stack.peek().path, ShouldEqual, "b")
		})

		Convey("calling pop should remove the last item on the stack", func() {
			stack.push(&stepperFile{
				fileInfo: &nilFileInfo{name: "a"},
				path:     "a",
			})
			stack.push(&stepperFile{
				fileInfo: &nilFileInfo{name: "b"},
				path:     "b",
			})

			So(stack.pop().path, ShouldEqual, "b")
			So(stack.peek().path, ShouldEqual, "a")
		})
	})
}

// stringStack tests

func TestStringStack(t *testing.T) {
	Convey("When creating a stringStack", t, func() {
		var stack = &stringStack{}

		Convey("calling clear should remove all items", func() {
			stack.push("a")
			stack.push("b")

			So(stack.isEmpty(), ShouldBeFalse)

			stack.clear()

			So(stack.isEmpty(), ShouldBeTrue)
		})

		Convey("calling peek should return the last item on the stack", func() {
			stack.push("a")
			stack.push("b")

			So(stack.peek(), ShouldEqual, "b")
		})

		Convey("calling pop should remove the last item on the stack", func() {
			stack.push("a")
			stack.push("b")

			So(stack.pop(), ShouldEqual, "b")
			So(stack.peek(), ShouldEqual, "a")
		})
	})
}

// Utility function tests

func TestFindFiles(t *testing.T) {
	Convey("When calling findFiles", t, func() {
		var err error

		Convey("it should return an error if the underlying filesystem returns an error when retrieving information "+
			"about the root path", func() {
			var stepper *pathStepper

			stepper, err = newPathStepper(&memFilesystem{
				statFileError: errors.New("statFile"),
			}, "/", true)

			So(err, ShouldNotBeNil)
			So(stepper, ShouldBeNil)

			So(err.Error(), ShouldEqual, "statFile")
		})

		Convey("it should return an error if the underlying filesystem returns an error when listing files", func() {
			var file File
			var root = &memFilesystemNode{
				children: map[string]*memFilesystemNode{
					"dir": {
						children: map[string]*memFilesystemNode{
							"file": {},
						},
					},
				},
			}
			var stepper *pathStepper

			stepper, err = newPathStepper(&memFilesystem{
				root: root,
			}, "/", true)

			So(err, ShouldBeNil)
			So(stepper, ShouldNotBeNil)

			stepper.fs = &memFilesystem{
				listFilesError: errors.New("listFiles"),
				root:           root,
			}

			file, err = stepper.nextFile()

			So(file, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "listFiles")
		})

		Convey("it should return an error if the underlying filesystem panics", func() {
			var fs Filesystem
			var root = &memFilesystemNode{
				children: map[string]*memFilesystemNode{
					"dir": {
						children: map[string]*memFilesystemNode{
							"file": {},
						},
					},
				},
			}

			fs = &memFilesystem{
				panic:         true,
				root:          root,
				statFileError: errors.New("statFile"),
			}

			err = findFiles(fs, "/", &stringStack{}, &stepperFileStack{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "a fatal error occurred: statFile")
		})
	})
}

func TestStripRoot(t *testing.T) {
	Convey("When calling stripRoot", t, func() {
		var result string

		Convey("with a path that starts with the path separator", func() {
			result = stripRoot("/abc", "/abc/xyz", "/")

			Convey("it should return the path with the root and the path separator stripped from it", func() {
				So(result, ShouldEqual, "xyz")
			})
		})

		Convey("with a path that does not start with the path separator", func() {
			result = stripRoot("/abc", "/abcxyz", "/")

			Convey("it should return the path with only the root stripped from it", func() {
				So(result, ShouldEqual, "xyz")
			})
		})
	})
}

func TestValidateID(t *testing.T) {
	Convey("When calling validateID", t, func() {
		Convey("it should succeed for valid IDs", func() {
			for _, id := range idsValid {
				So(validateID(id), ShouldBeNil)
			}
		})

		Convey("it should fail for invalid IDs", func() {
			for _, id := range idsInvalid {
				So(validateID(id), ShouldNotBeNil)
			}
		})
	})
}
