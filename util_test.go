package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"bytes"
	"errors"

	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Testcases
//

// cancellationHelper tests

var _ = g.Describe("cancellationHelper", func() {
	var buffer *bytes.Buffer
	var context Context
	var helper *cancellationHelper

	g.Describe("given a new instance", func() {
		g.BeforeEach(func() {
			buffer = new(bytes.Buffer)
			context = NewContext(ContextConfig{
				Writer: buffer,
			})
		})

		g.Describe("calling finalize", func() {
			g.Context("with a panic occurring inside finalize", func() {
				g.JustBeforeEach(func() {
					helper = newCancellationHelper(context.Log(), nil, nil, nil)
				})

				g.It("should log a warning", func() {
					helper.finalize()

					Expect(buffer.String()).To(ContainSubstring("an unexpected error occurred during cancellation"))
				})

				g.It("should still call the cancellation callback", func() {
					var value = 1

					helper.callback = func() {
						value = 2
					}

					helper.finalize()

					Expect(buffer.String()).To(ContainSubstring("an unexpected error occurred during cancellation"))
					Expect(value).To(Equal(2))
				})
			})

			g.Context("with a panic occurring inside the invoker's callback function", func() {
				g.JustBeforeEach(func() {
					helper = newCancellationHelper(context.Log(), make(chan Result), make(chan<- struct{}), nil)
				})

				g.It("should log a warning", func() {
					helper.callback = func() {
						panic("invoker panic")
					}

					helper.finalize()

					Expect(buffer.String()).To(ContainSubstring("invoker panic"))
				})
			})
		})
	})
})

// pathStepper tests

var _ = g.Describe("pathStepper", func() {
	var err error
	var stepper *pathStepper

	g.Describe("calling newPathStepper", func() {
		g.Context("with a Filesystem that throws an error when getting the absolute path", func() {
			g.BeforeEach(func() {
				stepper, err = newPathStepper(&memFilesystem{
					absolutePathError: errors.New("absolutePath"),
				}, "/", false)
			})

			g.It("should return an error", func() {
				Expect(stepper).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})

	g.Describe("given a new instance", func() {
		g.Describe("calling nextFile", func() {
			var file File

			g.Context("when a single file is the root", func() {
				g.BeforeEach(func() {
					stepper, err = newPathStepper(&memFilesystem{
						root: &memFilesystemNode{
							children: map[string]*memFilesystemNode{
								"file": {},
							},
						},
					}, "/file", false)
				})

				g.It("should return a single file", func() {
					Expect(err).To(BeNil())
					Expect(stepper).ToNot(BeNil())

					file, err = stepper.nextFile()

					Expect(err).To(BeNil())
					Expect(file).ToNot(BeNil())
					Expect(file.Path().String()).To(Equal("file"))

					file, err = stepper.nextFile()

					Expect(err).To(BeNil())
					Expect(file).To(BeNil())
				})
			})

			g.Context("when a nested directory structure is the root", func() {
				g.BeforeEach(func() {
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
				})

				g.It("should return the expected files", func() {
					Expect(err).To(BeNil())
					Expect(stepper).ToNot(BeNil())

					for i := 0; i < 3; i++ {
						file, err = stepper.nextFile()

						Expect(err).To(BeNil())
						Expect(file).ToNot(BeNil())
					}

					// And make sure calling it again will give us no results.

					file, err = stepper.nextFile()

					Expect(err).To(BeNil())
					Expect(file).To(BeNil())
				})
			})
		})
	})
})

// stepperFileStack tests

var _ = g.Describe("stepperFileStack", func() {
	g.Describe("given a new instance", func() {
		var stack *stepperFileStack

		g.BeforeEach(func() {
			stack = &stepperFileStack{}
		})

		g.Describe("calling clear", func() {
			g.It("should remove all items", func() {
				stack.push(&stepperFile{
					fileInfo: &nilFileInfo{name: "a"},
					path:     "a",
				})
				stack.push(&stepperFile{
					fileInfo: &nilFileInfo{name: "b"},
					path:     "b",
				})

				Expect(stack.isEmpty()).To(BeFalse())

				stack.clear()

				Expect(stack.isEmpty()).To(BeTrue())
			})
		})

		g.Describe("calling peek", func() {
			g.It("should return the last item on the stack", func() {
				stack.push(&stepperFile{
					fileInfo: &nilFileInfo{name: "a"},
					path:     "a",
				})
				stack.push(&stepperFile{
					fileInfo: &nilFileInfo{name: "b"},
					path:     "b",
				})

				Expect(stack.peek()).ToNot(BeNil())
				Expect(stack.peek().path).To(Equal("b"))
			})
		})

		g.Describe("calling pop", func() {
			g.It("should remove the last item on the stack", func() {
				var popped *stepperFile

				stack.push(&stepperFile{
					fileInfo: &nilFileInfo{name: "a"},
					path:     "a",
				})
				stack.push(&stepperFile{
					fileInfo: &nilFileInfo{name: "b"},
					path:     "b",
				})

				popped = stack.pop()

				Expect(popped).ToNot(BeNil())
				Expect(popped.path).To(Equal("b"))
				Expect(stack.peek()).ToNot(BeNil())
				Expect(stack.peek().path).To(Equal("a"))
			})
		})
	})
})

// stringStack tests

var _ = g.Describe("stringStack", func() {
	g.Describe("given a new instance", func() {
		var stack *stringStack

		g.BeforeEach(func() {
			stack = &stringStack{}
		})

		g.Describe("calling clear", func() {
			g.It("should remove all items", func() {
				stack.push("a")
				stack.push("b")

				Expect(stack.isEmpty()).To(BeFalse())

				stack.clear()

				Expect(stack.isEmpty()).To(BeTrue())
			})
		})

		g.Describe("calling peek", func() {
			g.It("should return the last item on the stack", func() {
				stack.push("a")
				stack.push("b")

				Expect(stack.peek()).ToNot(BeNil())
				Expect(stack.peek()).To(Equal("b"))
			})
		})

		g.Describe("calling pop", func() {
			g.It("should remove the last item on the stack", func() {
				stack.push("a")
				stack.push("b")

				Expect(stack.pop()).To(Equal("b"))
				Expect(stack.peek()).To(Equal("a"))
			})
		})
	})
})

// Utility function tests

var _ = g.Describe("findFiles", func() {
	g.Describe("calling findFiles", func() {
		var err error
		var root *memFilesystemNode

		g.BeforeEach(func() {
			root = &memFilesystemNode{
				children: map[string]*memFilesystemNode{
					"dir": {
						children: map[string]*memFilesystemNode{
							"file": {},
						},
					},
				},
			}
		})

		g.Context("with a Filesystem that returns an error when retrieving information about the root path", func() {
			var stepper *pathStepper

			g.JustBeforeEach(func() {
				stepper, err = newPathStepper(&memFilesystem{
					statFileError: errors.New("statFile"),
				}, "/", true)
			})

			g.It("should return an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(stepper).To(BeNil())

				Expect(err.Error()).To(Equal("statFile"))
			})
		})

		g.Context("with a Filesystem that returns an error when listing files", func() {
			var stepper *pathStepper

			g.JustBeforeEach(func() {
				stepper, err = newPathStepper(&memFilesystem{
					root: root,
				}, "/", true)
			})

			g.It("should return an error", func() {
				var file File

				Expect(err).To(BeNil())
				Expect(stepper).ToNot(BeNil())

				stepper.fs = &memFilesystem{
					listFilesError: errors.New("listFiles"),
					root:           root,
				}

				file, err = stepper.nextFile()

				Expect(err).ToNot(BeNil())
				Expect(file).To(BeNil())

				Expect(err.Error()).To(Equal("listFiles"))
			})
		})

		g.Context("with a Filesystem that panics at any point", func() {
			var fs *memFilesystem

			g.JustBeforeEach(func() {
				fs = &memFilesystem{
					panic:         true,
					root:          root,
					statFileError: errors.New("statFile"),
				}
			})

			g.It("should return an error", func() {
				err = findFiles(fs, "/", &stringStack{}, &stepperFileStack{})

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("a fatal error occurred: statFile"))
			})
		})
	})
})

var _ = g.Describe("stripRoot", func() {
	g.Describe("calling stripRoot", func() {
		var result string

		g.Context("with a path that starts with the path separator", func() {
			g.BeforeEach(func() {
				result = stripRoot("/abc", "/abc/xyz", "/")
			})

			g.It("should return the path with the root and the path separator stripped from it", func() {
				Expect(result).To(Equal("xyz"))
			})
		})

		g.Context("with a path that does not start with the path separator", func() {
			g.BeforeEach(func() {
				result = stripRoot("/abc", "/abcxyz", "/")
			})

			g.It("should return the path with onlyl the root stripped from it", func() {
				Expect(result).To(Equal("xyz"))
			})
		})
	})
})

var _ = g.Describe("validateID", func() {
	g.Describe("calling validateID", func() {
		g.Context("with valid IDs", func() {
			g.It("should succeed", func() {
				for _, id := range idsValid {
					Expect(validateID(id)).To(BeNil())
				}
			})
		})

		g.Context("with invalid IDs", func() {
			g.It("should fail", func() {
				for _, id := range idsInvalid {
					Expect(validateID(id)).ToNot(BeNil())
				}
			})
		})
	})
})

//
// Private variables
//

var (
	idsInvalid = []string{"", " ", ".", "a ", " a", "a.", ".a", "a..b", "a-b", "?"}
	idsValid   = []string{"a", "0", "a.0", "0.1", "a.b.c", "abc.def", "0.1.2", "0.abc.1.def"}
)
