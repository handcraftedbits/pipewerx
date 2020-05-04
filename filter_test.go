package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// Filter tests

func TestFilter(t *testing.T) {
	eventSinkTestMutex.Lock()
	defer eventSinkTestMutex.Unlock()

	Convey("When creating a Filter", t, func() {
		var err error
		var filter Filter
		var sink = &testEventSink{}
		var source Source

		Reset(resetGlobalEventSink)

		RegisterEventSink(sink)
		allowEventsFrom(componentFilter, true)

		Convey("which uses a FileEvaluator that discards some files and returns an error on destroy", func() {
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

			So(err, ShouldBeNil)
			So(source, ShouldNotBeNil)

			filter, err = NewFilter(FilterConfig{ID: "filter"}, []Source{source}, &extensionFileEvaluator{
				destroyError: errors.New("destroy"),
				extension:    "keep",
			})

			So(err, ShouldBeNil)
			So(filter, ShouldNotBeNil)

			Convey("calling destroy should return the expected value", func() {
				err = filter.destroy()

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "destroy")

				Convey("and the appropriate event should be sent", func() {
					sink.expectEvents(eventFilterCreated, eventFilterDestroyed)
				})
			})

			Convey("calling Files", func() {
				var results []Result

				Convey("should return the expected Results", func() {
					results = collectSourceResults(filter)

					expectFilePathsInResults(nil, results, []string{"file1.keep", "file2.keep", "file3.keep",
						"file4.keep"})

					Convey("and the appropriate events should be sent", func() {
						sink.expectEvents(eventFilterCreated, eventFilterStarted, eventFilterResultProduced,
							eventFilterResultProduced, eventFilterResultProduced, eventFilterResultProduced,
							eventFilterFinished)
					})
				})

				Convey("and cancelling should return the expected Results", func(c C) {
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

						c.So(results, ShouldHaveLength, 1)

						expectFilePathsInResults(c, results, []string{"file1.keep", "file2.keep", "file3.keep",
							"file4.keep"})

						wg.Done()
					})

					wg.Wait()

					Convey("and the appropriate events should be sent", func() {
						sink.expectEvents(eventFilterCreated, eventFilterStarted, eventFilterResultProduced,
							eventFilterCancelled, eventFilterFinished)
					})
				})
			})

			Convey("calling ID should return the expected ID", func() {
				So(filter.ID(), ShouldEqual, "filter")
			})
		})

		Convey("which uses a FileEvaluator that panics when calling ShouldKeep", func() {
			source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{
				root: &memFilesystemNode{
					children: map[string]*memFilesystemNode{
						"file1.keep": {},
					},
				},
			})

			So(err, ShouldBeNil)
			So(source, ShouldNotBeNil)

			filter, err = NewFilter(FilterConfig{ID: "filter"}, []Source{source}, &extensionFileEvaluator{
				extension:       "keep",
				panic:           true,
				shouldKeepError: errors.New("shouldKeep"),
			})

			Convey("calling Files should return an error Result", func() {
				var results = collectSourceResults(filter)

				So(results, ShouldHaveLength, 1)
				So(results[0].File(), ShouldBeNil)
				So(results[0].Error(), ShouldNotBeNil)
				So(results[0].Error().Error(), ShouldEqual, "a fatal error occurred: shouldKeep")

				Convey("and the appropriate events should be sent", func() {
					sink.expectEvents(eventFilterCreated, eventFilterStarted, eventFilterResultProduced,
						eventFilterFinished)
				})
			})
		})
	})
}

func TestFilterEvents(t *testing.T) {
	var id = "filter"

	Convey("When calling filterEventCancelled", t, func() {
		validateSourceEvent(filterEventCancelled(id), componentFilter, eventTypeCancelled, id)
	})

	Convey("When calling filterEventCreated", t, func() {
		validateSourceEvent(filterEventCreated(id), componentFilter, eventTypeCreated, id)
	})

	Convey("When calling filterEventDestroyed", t, func() {
		validateSourceEvent(filterEventDestroyed(id), componentFilter, eventTypeDestroyed, id)
	})

	Convey("When calling filterEventFinished", t, func() {
		validateSourceEvent(filterEventFinished(id), componentFilter, eventTypeFinished, id)
	})

	Convey("When calling filterEventResultProduced", t, func() {
		var res = &result{
			err: errors.New("result error"),
			file: &file{
				fileInfo: &nilFileInfo{
					name: "name",
				},
				path: newFilePath(nil, "name", "/"),
			},
		}

		validateSourceEvent(filterEventResultProduced(id, res), componentFilter, eventTypeResultProduced, id)
	})

	Convey("When calling filterEventStarted", t, func() {
		validateSourceEvent(filterEventStarted(id), componentFilter, eventTypeStarted, id)
	})
}

func TestNewFilter(t *testing.T) {
	eventSinkTestMutex.Lock()
	defer eventSinkTestMutex.Unlock()

	Convey("When calling NewFilter", t, func() {
		var err error
		var filter Filter
		var sink = &testEventSink{}

		Reset(resetGlobalEventSink)

		RegisterEventSink(sink)
		allowEventsFrom(componentFilter, true)

		Convey("it should succeed for valid IDs", func() {
			var source Source

			source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{})

			So(err, ShouldBeNil)
			So(source, ShouldNotBeNil)

			for _, id := range idsValid {
				filter, err = NewFilter(FilterConfig{ID: id}, []Source{source}, nil)

				So(err, ShouldBeNil)
				So(filter, ShouldNotBeNil)
			}
		})

		Convey("it should fail for invalid IDs", func() {
			for _, id := range idsInvalid {
				filter, err = NewFilter(FilterConfig{ID: id}, nil, nil)

				So(filter, ShouldBeNil)
				So(err, ShouldNotBeNil)
			}
		})

		Convey("with a nil Source array", func() {
			filter, err = NewFilter(FilterConfig{ID: "filter"}, nil, nil)

			Convey("it should return an error", func() {
				So(filter, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errSourceNone), ShouldBeTrue)
			})
		})

		Convey("with an empty Source array", func() {
			filter, err = NewFilter(FilterConfig{ID: "filter"}, []Source{}, nil)

			Convey("it should return an error", func() {
				So(filter, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errSourceNone), ShouldBeTrue)
			})
		})

		Convey("with a nil FileEvaluator", func() {
			var source Source

			source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{
				root: &memFilesystemNode{
					children: map[string]*memFilesystemNode{
						"file1": {},
						"file2": {},
					},
				},
			})

			So(err, ShouldBeNil)
			So(source, ShouldNotBeNil)

			filter, err = NewFilter(FilterConfig{ID: "filter"}, []Source{source}, nil)

			So(err, ShouldBeNil)
			So(filter, ShouldNotBeNil)

			Convey("calling Destroy should not perform any action", func() {
				So(filter.destroy(), ShouldBeNil)

				Convey("and the appropriate event should be sent", func() {
					sink.expectEvents(eventFilterCreated, eventFilterDestroyed)
				})
			})

			Convey("calling Files should return all files", func() {
				var results = collectSourceResults(filter)

				So(results, ShouldHaveLength, 2)

				expectFilePathsInResults(nil, results, []string{"file1", "file2"})

				Convey("and the appropriate events should be sent", func() {
					sink.expectEvents(eventFilterCreated, eventFilterStarted, eventFilterResultProduced,
						eventFilterResultProduced, eventFilterFinished)
				})
			})
		})
	})
}
