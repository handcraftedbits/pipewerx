package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"
	"sync"

	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Testcases
//

// Filter tests

var _ = g.Describe("Filter", func() {
	g.Describe("given a new instance", func() {
		var err error
		var filter Filter
		var sink *testEventSink
		var source Source

		g.BeforeEach(func() {
			sink = newTestEventSink()

			RegisterEventSink(sink)
		})

		g.Context("which uses a FileEvaluator that discards some files and returns an error on destroy", func() {
			g.JustBeforeEach(func() {
				source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{
					root: &memFilesystemNode{
						children: map[string]*memFilesystemNode{
							"file1.keep":   {},
							"file1.nokeep": {},
							"file2.keep":   {},
							"file2.nokeep": {},
							"file3.keep":   {},
							"file3.nokeep": {},
							"file4.keep":   {},
							"file4.nokeep": {},
						},
					},
				})

				Expect(err).To(BeNil())
				Expect(source).NotTo(BeNil())

				filter, err = NewFilter(FilterConfig{ID: sink.id}, []Source{source}, &extensionFileEvaluator{
					destroyError: errors.New("destroy"),
					extension:    "keep",
				})

				Expect(err).To(BeNil())
				Expect(filter).NotTo(BeNil())
			})

			g.Describe("calling destroy", func() {
				g.It("should not return an error and it should send the appropriate event", func() {
					err = filter.destroy()

					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(Equal("destroy"))

					Expect(sink).To(haveTheseEvents(eventFilterCreated, eventFilterDestroyed))
				})
			})

			g.Describe("calling Files", func() {
				var results []Result

				g.Context("without cancelling", func() {
					g.It("should return the expected Results and send the appropriate events", func() {
						results = collectSourceResults(filter)

						Expect(results).To(HaveLen(4))
						Expect(results).To(haveAllOrSomeOfTheseFilePaths("file1.keep", "file2.keep", "file3.keep",
							"file4.keep"))

						Expect(sink).To(haveTheseEvents(eventFilterCreated, eventFilterStarted,
							eventFilterResultProduced, eventFilterResultProduced, eventFilterResultProduced,
							eventFilterResultProduced, eventFilterFinished))
					})
				})

				g.Context("and cancelling", func() {
					g.It("should return the expected Results and send the appropriate events", func() {
						var cancel CancelFunc
						var in <-chan Result
						var wg sync.WaitGroup

						results = make([]Result, 0)

						in, cancel = filter.Files(NewContext(ContextConfig{}))

						results = append(results, <-in)

						wg.Add(1)

						cancel(func() {
							var names []string

							for result := range in {
								results = append(results, result)

								names = append(names, result.File().Path().String())
							}

							Expect(results).To(HaveLen(1))
							Expect(results).To(haveAllOrSomeOfTheseFilePaths("file1.keep", "file2.keep", "file3.keep",
								"file4.keep"))

							wg.Done()
						})

						wg.Wait()

						Expect(sink).To(haveTheseEvents(eventFilterCreated, eventFilterStarted,
							eventFilterResultProduced, eventFilterCancelled, eventFilterFinished))
					})
				})
			})
		})

		g.Context("which uses a FileEvaluator that panics when calling ShouldKeep", func() {
			g.JustBeforeEach(func() {
				source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{
					root: &memFilesystemNode{
						children: map[string]*memFilesystemNode{
							"file1.keep": {},
						},
					},
				})

				Expect(err).To(BeNil())
				Expect(source).NotTo(BeNil())

				filter, err = NewFilter(FilterConfig{ID: sink.id}, []Source{source}, &extensionFileEvaluator{
					extension:       "keep",
					panic:           true,
					shouldKeepError: errors.New("shouldKeep"),
				})

				Expect(err).To(BeNil())
				Expect(filter).NotTo(BeNil())
			})

			g.Describe("calling Files", func() {
				g.It("should return an error Result and send the appropriate events", func() {
					var results = collectSourceResults(filter)

					Expect(results).To(HaveLen(1))
					Expect(results[0].File()).To(BeNil())
					Expect(results[0].Error()).NotTo(BeNil())
					Expect(results[0].Error().Error()).To(Equal("a fatal error occurred: shouldKeep"))

					Expect(sink).To(haveTheseEvents(eventFilterCreated, eventFilterStarted, eventFilterResultProduced,
						eventFilterFinished))
				})
			})
		})
	})
})

var _ = g.Describe("Filter events", func() {
	var id = "filter"

	g.Describe("calling filterEventCancelled", func() {
		g.It("should return a valid event", func() {
			Expect(filterEventCancelled(id)).To(beAValidEvent(componentFilter, eventTypeCancelled, id))
		})
	})

	g.Describe("calling filterEventCreated", func() {
		g.It("should return a valid event", func() {
			Expect(filterEventCreated(id)).To(beAValidEvent(componentFilter, eventTypeCreated, id))
		})
	})

	g.Describe("calling filterEventDestroyed", func() {
		g.It("should return a valid event", func() {
			Expect(filterEventDestroyed(id)).To(beAValidEvent(componentFilter, eventTypeDestroyed, id))
		})
	})

	g.Describe("calling filterEventFinished", func() {
		g.It("should return a valid event", func() {
			Expect(filterEventFinished(id)).To(beAValidEvent(componentFilter, eventTypeFinished, id))
		})
	})

	g.Describe("calling filterEventResultProduced", func() {
		g.It("should return a valid event", func() {
			var res = &result{
				err: errors.New("result error"),
				file: &file{
					fileInfo: &nilFileInfo{
						name: "name",
					},
					path: newFilePath(nil, "name", "/"),
				},
			}

			Expect(filterEventResultProduced(id, res)).To(beAValidEvent(componentFilter, eventTypeResultProduced, id))
		})
	})

	g.Describe("calling filterEventStarted", func() {
		g.It("should return a valid event", func() {
			Expect(filterEventStarted(id)).To(beAValidEvent(componentFilter, eventTypeStarted, id))
		})
	})
})

var _ = g.Describe("NewFilter", func() {
	g.Describe("calling NewFilter", func() {
		var err error
		var filter Filter
		var source Source

		g.BeforeEach(func() {
			source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{})

			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())
		})

		g.Context("with valid IDs", func() {
			g.It("should succeed", func() {
				for _, id := range idsValid {
					filter, err = NewFilter(FilterConfig{ID: id}, []Source{source}, nil)

					Expect(err).To(BeNil())
					Expect(filter).NotTo(BeNil())
				}
			})
		})

		g.Context("with invalid IDs", func() {
			g.It("should return an error", func() {
				for _, id := range idsInvalid {
					filter, err = NewFilter(FilterConfig{ID: id}, []Source{source}, nil)

					Expect(filter).To(BeNil())
					Expect(err).NotTo(BeNil())
				}
			})
		})

		g.Context("with a nil Source array", func() {
			g.It("should return an error", func() {
				filter, err = NewFilter(FilterConfig{ID: "filter"}, nil, nil)

				Expect(filter).To(BeNil())
				Expect(err).NotTo(BeNil())
				Expect(errors.Is(err, errSourceNone)).To(BeTrue())
			})
		})

		g.Context("with an empty Source array", func() {
			g.It("should return an error", func() {
				filter, err = NewFilter(FilterConfig{ID: "filter"}, []Source{}, nil)

				Expect(filter).To(BeNil())
				Expect(err).NotTo(BeNil())
				Expect(errors.Is(err, errSourceNone)).To(BeTrue())
			})
		})

		g.Context("with a nil FileEvaluator", func() {
			var sink *testEventSink

			g.BeforeEach(func() {
				sink = newTestEventSink()

				RegisterEventSink(sink)

				source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{
					root: &memFilesystemNode{
						children: map[string]*memFilesystemNode{
							"file1": {},
							"file2": {},
						},
					},
				})

				Expect(err).To(BeNil())
				Expect(source).NotTo(BeNil())

				filter, err = NewFilter(FilterConfig{ID: sink.id}, []Source{source}, nil)

				Expect(err).To(BeNil())
				Expect(filter).NotTo(BeNil())
			})

			g.Describe("calling destroy", func() {
				g.It("should not perform any action and it should send the appropriate events", func() {
					Expect(filter.destroy()).To(BeNil())

					Expect(sink).To(haveTheseEvents(eventFilterCreated, eventFilterDestroyed))
				})
			})

			g.Describe("calling Files", func() {
				g.It("should return all files and send the appropriate events", func() {
					var results = collectSourceResults(filter)

					Expect(results).To(HaveLen(2))
					Expect(results).To(haveAllOrSomeOfTheseFilePaths("file1", "file2"))

					Expect(sink).To(haveTheseEvents(eventFilterCreated, eventFilterStarted, eventFilterResultProduced,
						eventFilterResultProduced, eventFilterFinished))
				})
			})
		})
	})
})

//
// Private constants
//

const (
	// Filter event names for testEventSink.expectEvents()
	eventFilterCancelled      = componentFilter + "." + eventTypeCancelled
	eventFilterCreated        = componentFilter + "." + eventTypeCreated
	eventFilterDestroyed      = componentFilter + "." + eventTypeDestroyed
	eventFilterFinished       = componentFilter + "." + eventTypeFinished
	eventFilterResultProduced = componentFilter + "." + eventTypeResultProduced
	eventFilterStarted        = componentFilter + "." + eventTypeStarted
)
