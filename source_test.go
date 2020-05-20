package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"
	"sync"

	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.handcraftedbits.com/pipewerx/internal/event"
)

//
// Testcases
//

// mergedSource tests

var _ = g.Describe("mergedSource", func() {
	g.Describe("given a new instance", func() {
		var err error
		var merged Source

		g.Context("which contains a number of Sources", func() {
			g.Context("each of which have a Filesystem that does nothing when destroyed", func() {
				g.BeforeEach(func() {
					var source [2]Source

					source[0], err = NewSource(SourceConfig{ID: "source0"}, &memFilesystem{
						root: &memFilesystemNode{
							children: map[string]*memFilesystemNode{
								"file1": {},
							},
						},
					})

					Expect(err).To(BeNil())
					Expect(source[0]).NotTo(BeNil())

					source[1], err = NewSource(SourceConfig{ID: "source1"}, &memFilesystem{
						root: &memFilesystemNode{
							children: map[string]*memFilesystemNode{
								"file2": {},
							},
						},
					})

					Expect(err).To(BeNil())
					Expect(source[1]).NotTo(BeNil())

					merged, err = newMergedSource("merged", []Source{source[0], source[1]})

					Expect(err).To(BeNil())
					Expect(merged).NotTo(BeNil())
				})

				g.Describe("calling destroy", func() {
					g.It("should return nil", func() {
						Expect(merged.destroy()).To(BeNil())
					})
				})
			})

			g.Context("most of which have a Filesystem that performs an action when destroyed", func() {
				var expected = []string{"file01", "file02", "file03", "file04", "file05", "file06", "file07", "file08",
					"file09", "file10", "file11", "file12", "file13", "file14", "file15"}

				g.BeforeEach(func() {
					var source [3]Source

					source[0], err = NewSource(SourceConfig{
						ID: "source1",
					}, &memFilesystem{
						destroy: func() error {
							return errors.New("source1")
						},
						root: &memFilesystemNode{
							children: map[string]*memFilesystemNode{
								"file01": {},
								"file02": {},
								"file03": {},
								"file04": {},
								"file05": {},
							},
						},
					})

					Expect(err).To(BeNil())
					Expect(source[0]).NotTo(BeNil())

					source[1], err = NewSource(SourceConfig{
						ID: "source2",
					}, &memFilesystem{
						destroy: func() error {
							return errors.New("source2")
						},
						root: &memFilesystemNode{
							children: map[string]*memFilesystemNode{
								"file06": {},
								"file07": {},
								"file08": {},
								"file09": {},
								"file10": {},
							},
						},
					})

					Expect(err).To(BeNil())
					Expect(source[1]).NotTo(BeNil())

					source[2], err = NewSource(SourceConfig{
						ID: "source3",
					}, &memFilesystem{
						root: &memFilesystemNode{
							children: map[string]*memFilesystemNode{
								"file11": {},
								"file12": {},
								"file13": {},
								"file14": {},
								"file15": {},
							},
						},
					})

					Expect(err).To(BeNil())
					Expect(source[2]).NotTo(BeNil())

					merged, err = newMergedSource("merged", []Source{source[0], source[1], source[2]})

					Expect(err).To(BeNil())
					Expect(merged).NotTo(BeNil())
				})

				g.Describe("calling destroy", func() {
					g.It("should return multiple errors", func() {
						var multiErr MultiError
						var ok bool

						err = merged.destroy()

						Expect(err).NotTo(BeNil())

						multiErr, ok = err.(MultiError)

						Expect(ok).To(BeTrue())

						Expect(multiErr.Causes()).NotTo(BeNil())
						Expect(multiErr.Causes()).To(HaveLen(2))
						Expect(multiErr.Causes()[0].Error()).To(Equal("source1"))
						Expect(multiErr.Causes()[1].Error()).To(Equal("source2"))
					})
				})

				g.Describe("calling Files", func() {
					var results []Result

					g.Context("without cancelling", func() {
						g.It("should return the expected Results", func() {
							results = collectSourceResults(merged)

							Expect(results).To(HaveLen(15))
							Expect(results).To(haveAllOrSomeOfTheseFilePaths(expected...))
						})
					})

					g.Context("and cancelling", func() {
						g.It("should return the expected Results", func() {
							var cancel CancelFunc
							var in <-chan Result
							var wg sync.WaitGroup

							results = make([]Result, 0)

							in, cancel = merged.Files(NewContext(ContextConfig{}))

							results = append(results, <-in)
							results = append(results, <-in)

							wg.Add(1)

							cancel(func() {
								var names []string

								for result := range in {
									results = append(results, result)

									names = append(names, result.File().Path().String())
								}

								Expect(results).To(HaveLen(2))
								Expect(results).To(haveAllOrSomeOfTheseFilePaths(expected...))

								wg.Done()
							})

							wg.Wait()
						})
					})
				})

				g.Describe("calling ID", func() {
					g.It("should return the expected ID", func() {
						Expect(merged.ID()).To(Equal("merged"))
					})
				})
			})
		})
	})
})

var _ = g.Describe("newMergedSource", func() {
	g.Describe("when calling newMergedSource", func() {
		var err error
		var merged Source

		g.Context("with a nil Source array", func() {
			g.It("should return an error", func() {
				merged, err = newMergedSource("merged", nil)

				Expect(merged).To(BeNil())
				Expect(err).NotTo(BeNil())
				Expect(errors.Is(err, errSourceNone)).To(BeTrue())
			})
		})

		g.Context("with an empty Source array", func() {
			g.It("should return an error", func() {
				merged, err = newMergedSource("merged", []Source{})

				Expect(merged).To(BeNil())
				Expect(err).NotTo(BeNil())
				Expect(errors.Is(err, errSourceNone)).To(BeTrue())
			})
		})

		g.Context("with a single Source", func() {
			var source Source

			g.BeforeEach(func() {
				source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{})

				Expect(err).To(BeNil())
				Expect(source).NotTo(BeNil())
			})

			g.It("should return the same Source", func() {
				merged, err = newMergedSource("merged", []Source{source})

				Expect(err).To(BeNil())
				Expect(merged).NotTo(BeNil())
				Expect(merged).To(Equal(source))
			})
		})

		g.Context("with duplicate Sources", func() {
			var source1 Source
			var source2 Source

			g.BeforeEach(func() {
				source1, err = NewSource(SourceConfig{ID: "source1"}, &memFilesystem{
					root: &memFilesystemNode{
						children: map[string]*memFilesystemNode{
							"file1": {},
						},
					},
				})

				Expect(err).To(BeNil())
				Expect(source1).NotTo(BeNil())

				source2, err = NewSource(SourceConfig{ID: "source2"}, &memFilesystem{
					root: &memFilesystemNode{
						children: map[string]*memFilesystemNode{
							"file2": {},
						},
					},
				})

				Expect(err).To(BeNil())
				Expect(source2).NotTo(BeNil())
			})

			g.It("should discard the duplicates", func() {
				var results []Result

				merged, err = newMergedSource("merged", []Source{source1, source2, source1, source2})

				Expect(err).To(BeNil())
				Expect(merged).NotTo(BeNil())

				results = collectSourceResults(merged)

				Expect(results).To(HaveLen(2))
				Expect(results).To(haveAllOrSomeOfTheseFilePaths("file1", "file2"))
			})
		})
	})
})

// Source tests

var _ = g.Describe("NewSource", func() {
	g.Describe("calling NewSource", func() {
		var err error
		var source Source

		g.Context("with valid IDs", func() {
			g.It("should succeed", func() {
				for _, id := range idsValid {
					source, err = NewSource(SourceConfig{ID: id}, &memFilesystem{})

					Expect(err).To(BeNil())
					Expect(source).ToNot(BeNil())
				}
			})
		})

		g.Context("with invalid IDs", func() {
			g.It("should return an error", func() {
				for _, id := range idsInvalid {
					source, err = NewSource(SourceConfig{ID: id}, &memFilesystem{})

					Expect(source).To(BeNil())
					Expect(err).NotTo(BeNil())
				}
			})
		})

		g.Context("with a nil Filesystem", func() {
			g.BeforeEach(func() {
				source, err = NewSource(SourceConfig{}, nil)

				Expect(source).To(BeNil())
				Expect(err).NotTo(BeNil())
			})

			g.It("should return an error", func() {
				Expect(errors.Is(err, errSourceNilFilesystem)).To(BeTrue())
			})
		})

		g.Context("with a valid Filesystem", func() {
			var sink *testEventSink

			g.BeforeEach(func() {
				sink = newTestEventSink()

				event.RegisterSink(sink)

				source, err = NewSource(SourceConfig{ID: sink.id}, &memFilesystem{})

				Expect(err).To(BeNil())
				Expect(source).ToNot(BeNil())
			})

			g.It("should not return an error and it should send the appropriate events", func() {
				Expect(sink).To(haveTheseEvents(eventSourceCreated))
			})
		})
	})
})

var _ = g.Describe("Source", func() {
	g.Describe("given a new instance", func() {
		var err error
		var sink *testEventSink
		var source Source

		g.BeforeEach(func() {
			sink = newTestEventSink()

			event.RegisterSink(sink)
		})

		g.Context("which fails immediately upon access and has a Filesystem which performs no action when destroyed",
			func() {
				g.JustBeforeEach(func() {
					source, err = NewSource(SourceConfig{ID: sink.id}, &memFilesystem{
						absolutePathError: errors.New("absolutePath"),
					})

					Expect(err).To(BeNil())
					Expect(source).NotTo(BeNil())
				})

				g.Describe("calling destroy", func() {
					g.It("should not perform any action and it should send the appropriate events", func() {
						Expect(source.destroy()).To(BeNil())

						Expect(sink).To(haveTheseEvents(eventSourceCreated, eventSourceDestroyed))
					})
				})

				g.Describe("calling Files", func() {
					g.It("should return an error and send the appropriate events", func() {
						var results = collectSourceResults(source)

						Expect(results).To(HaveLen(1))
						Expect(results[0].File()).To(BeNil())
						Expect(results[0].Error()).NotTo(BeNil())
						Expect(results[0].Error().Error()).To(Equal("absolutePath"))

						Expect(sink).To(haveTheseEvents(eventSourceCreated, eventSourceStarted,
							eventSourceResultProduced, eventSourceFinished))
					})
				})
			})

		g.Context("which contains a number of files and has a Filesystem which performs an action when destroyed",
			func() {
				g.JustBeforeEach(func() {
					source, err = NewSource(SourceConfig{
						ID:      sink.id,
						Recurse: true,
					}, &memFilesystem{
						destroy: func() error {
							return errors.New("destroy")
						},
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
								"file": {},
							},
						},
					})

					Expect(err).To(BeNil())
					Expect(source).NotTo(BeNil())
				})

				g.Describe("calling destroy", func() {
					g.It("should return an error and send the appropriate events", func() {
						err = source.destroy()

						Expect(err).NotTo(BeNil())
						Expect(err.Error()).To(Equal("destroy"))

						Expect(sink).To(haveTheseEvents(eventSourceCreated, eventSourceDestroyed))
					})
				})

				g.Describe("calling Files", func() {
					var results []Result

					g.Context("without cancelling", func() {
						g.It("should return the expected Results and send the appropriate events", func() {
							results = collectSourceResults(source)

							Expect(results).To(HaveLen(3))
							Expect(results).To(haveAllOrSomeOfTheseFilePaths("file", "dir1/file1", "dir2/file2"))

							Expect(sink).To(haveTheseEvents(eventSourceCreated, eventSourceStarted,
								eventSourceResultProduced, eventSourceResultProduced, eventSourceResultProduced,
								eventSourceFinished))
						})
					})

					g.Context("and cancelling", func() {
						g.It("should return the expected Results and send the appropriate events", func() {
							var cancel CancelFunc
							var in <-chan Result
							var wg sync.WaitGroup

							results = make([]Result, 0)

							in, cancel = source.Files(NewContext(ContextConfig{}))

							results = append(results, <-in)
							results = append(results, <-in)

							wg.Add(1)

							cancel(func() {
								for result := range in {
									results = append(results, result)
								}

								Expect(results).To(HaveLen(2))
								Expect(results).To(haveAllOrSomeOfTheseFilePaths("file", "dir1/file1", "dir2/file2"))

								wg.Done()
							})

							wg.Wait()

							Expect(sink).To(haveTheseEvents(eventSourceCreated, eventSourceStarted,
								eventSourceResultProduced, eventSourceResultProduced, eventSourceCancelled,
								eventSourceFinished))
						})
					})
				})

				g.Describe("calling ID", func() {
					g.It("should return the expected ID", func() {
						Expect(source.ID()).To(Equal(sink.id))
					})
				})
			})

		g.Context("which panics when calling Files", func() {
			g.JustBeforeEach(func() {
				source, err = NewSource(SourceConfig{ID: sink.id}, &memFilesystem{
					absolutePathError: errors.New("absolutePath"),
					panic:             true,
				})

				Expect(err).To(BeNil())
				Expect(source).NotTo(BeNil())
			})

			g.Describe("calling Files", func() {
				var results []Result

				g.It("should return an error instead of panicking and send the appropriate events", func() {
					results = collectSourceResults(source)

					Expect(results).To(HaveLen(1))
					Expect(results[0].File()).To(BeNil())
					Expect(results[0].Error()).NotTo(BeNil())
					Expect(results[0].Error().Error()).To(Equal("a fatal error occurred: absolutePath"))

					Expect(sink).To(haveTheseEvents(eventSourceCreated, eventSourceStarted, eventSourceResultProduced,
						eventSourceFinished))
				})
			})
		})
	})
})

//
// Private constants
//

const (
	// Source event names for testEventSink.expectEvents()
	eventSourceCancelled      = componentSource + "." + event.TypeCancelled
	eventSourceCreated        = componentSource + "." + event.TypeCreated
	eventSourceDestroyed      = componentSource + "." + event.TypeDestroyed
	eventSourceFinished       = componentSource + "." + event.TypeFinished
	eventSourceResultProduced = componentSource + "." + event.TypeResultProduced
	eventSourceStarted        = componentSource + "." + event.TypeStarted
)
