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

// mergedSource tests

func TestMergedSource(t *testing.T) {
	Convey("When creating a mergedSource", t, func() {
		var err error
		var merged Source

		Convey("which contains a number of Sources, each of which have a Filesystem that does nothing when destroyed",
			func() {
				var source [2]Source

				source[0], err = NewSource(SourceConfig{ID: "source0"}, &memFilesystem{
					root: &memFilesystemNode{
						children: map[string]*memFilesystemNode{
							"file1": {},
						},
					},
				})

				So(err, ShouldBeNil)
				So(source[0], ShouldNotBeNil)

				source[1], err = NewSource(SourceConfig{ID: "source1"}, &memFilesystem{
					root: &memFilesystemNode{
						children: map[string]*memFilesystemNode{
							"file2": {},
						},
					},
				})

				So(err, ShouldBeNil)
				So(source[1], ShouldNotBeNil)

				merged, err = newMergedSource("merged", []Source{source[0], source[1]})

				So(err, ShouldBeNil)
				So(merged, ShouldNotBeNil)

				Convey("calling destroy should return nil", func() {
					So(merged.destroy(), ShouldEqual, nil)
				})
			})

		Convey("which contains a number of Sources, most of which have a Filesystem that performs an action when "+
			"destroyed", func() {
			var expected = []string{"file01", "file02", "file03", "file04", "file05", "file06", "file07", "file08",
				"file09", "file10", "file11", "file12", "file13", "file14", "file15"}
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

			So(err, ShouldBeNil)
			So(source[0], ShouldNotBeNil)

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

			So(err, ShouldBeNil)
			So(source[1], ShouldNotBeNil)

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

			So(err, ShouldBeNil)
			So(source[2], ShouldNotBeNil)

			merged, err = newMergedSource("merged", []Source{source[0], source[1], source[2]})

			So(err, ShouldBeNil)
			So(merged, ShouldNotBeNil)

			Convey("calling destroy should return the expected value", func() {
				var multiErr MultiError

				err = merged.destroy()

				So(err, ShouldNotBeNil)
				So(err, ShouldImplement, (*MultiError)(nil))

				multiErr = err.(MultiError)

				So(multiErr.Causes(), ShouldHaveLength, 2)
				So(multiErr.Causes()[0].Error(), ShouldEqual, "source1")
				So(multiErr.Causes()[1].Error(), ShouldEqual, "source2")
			})

			Convey("calling Files", func() {
				var results []Result

				Convey("should return the expected Results", func() {
					results = collectSourceResults(merged)

					So(results, ShouldHaveLength, 15)

					expectFilePathsInResults(nil, results, expected)
				})

				Convey("and cancelling should return the expected Results", func(c C) {
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

						c.So(results, ShouldHaveLength, 2)

						expectFilePathsInResults(c, results, expected)

						wg.Done()
					})

					wg.Wait()
				})
			})

			Convey("calling ID should return the expected ID", func() {
				So(merged.ID(), ShouldEqual, "merged")
			})
		})
	})
}

func TestNewMergedSource(t *testing.T) {
	Convey("When calling newMergedSource", t, func() {
		var err error
		var merged Source

		Convey("with a nil Source array", func() {
			merged, err = newMergedSource("merged", nil)

			Convey("it should return an error", func() {
				So(merged, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errSourceNone), ShouldBeTrue)
			})
		})

		Convey("with an empty Source array", func() {
			merged, err = newMergedSource("merged", []Source{})

			Convey("it should return an error", func() {
				So(merged, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errSourceNone), ShouldBeTrue)
			})
		})

		Convey("with a single Source", func() {
			var source Source

			source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{})

			So(err, ShouldBeNil)
			So(source, ShouldNotBeNil)

			merged, err = newMergedSource("merged", []Source{source})

			Convey("it should return the same Source", func() {
				So(err, ShouldBeNil)
				So(merged, ShouldNotBeNil)
				So(merged, ShouldEqual, source)
			})
		})

		Convey("with duplicate Sources", func() {
			var source1 Source
			var source2 Source

			source1, err = NewSource(SourceConfig{ID: "source1"}, &memFilesystem{
				root: &memFilesystemNode{
					children: map[string]*memFilesystemNode{
						"file1": {},
					},
				},
			})

			So(err, ShouldBeNil)
			So(source1, ShouldNotBeNil)

			source2, err = NewSource(SourceConfig{ID: "source2"}, &memFilesystem{
				root: &memFilesystemNode{
					children: map[string]*memFilesystemNode{
						"file2": {},
					},
				},
			})

			So(err, ShouldBeNil)
			So(source2, ShouldNotBeNil)

			merged, err = newMergedSource("merged", []Source{source1, source2, source1, source2})

			Convey("it should discard the duplicates", func() {
				var results []Result

				So(err, ShouldBeNil)
				So(merged, ShouldNotBeNil)

				results = collectSourceResults(merged)

				So(results, ShouldHaveLength, 2)

				expectFilePathsInResults(nil, results, []string{"file1", "file2"})
			})
		})
	})
}

// Source tests

func TestNewSource(t *testing.T) {
	eventSinkTestMutex.Lock()
	defer eventSinkTestMutex.Unlock()

	Convey("When calling NewSource", t, func() {
		var err error
		var source Source

		Convey("it should succeed for valid IDs", func() {
			for _, id := range idsValid {
				source, err = NewSource(SourceConfig{ID: id}, &memFilesystem{})

				So(err, ShouldBeNil)
				So(source, ShouldNotBeNil)
			}
		})

		Convey("it should fail for invalid IDs", func() {
			for _, id := range idsInvalid {
				source, err = NewSource(SourceConfig{ID: id}, &memFilesystem{})

				So(source, ShouldBeNil)
				So(err, ShouldNotBeNil)
			}
		})

		Convey("with a nil Filesystem", func() {
			Convey("it should return an error", func() {
				source, err = NewSource(SourceConfig{}, nil)

				So(source, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errSourceNilFilesystem), ShouldBeTrue)
			})
		})

		Convey("with a valid Filesystem", func() {
			Reset(resetGlobalEventSink)

			Convey("it should not return an error", func() {
				var sink = &testEventSink{}

				RegisterEventSink(sink)
				allowEventsFrom(componentSource, true)

				source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{})

				So(err, ShouldBeNil)
				So(source, ShouldNotBeNil)

				Convey("and the appropriate event should be sent", func() {
					sink.expectEvents(eventSourceCreated)
				})
			})
		})
	})
}

func TestSource(t *testing.T) {
	eventSinkTestMutex.Lock()
	eventSinkTestMutex.Unlock()

	Convey("When creating a Source", t, func() {
		var err error
		var sink = &testEventSink{}
		var source Source

		Reset(resetGlobalEventSink)

		RegisterEventSink(sink)
		allowEventsFrom(componentSource, true)

		Convey("which fails immediately upon access and has a Filesystem which performs no action when destroyed",
			func() {
				source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{
					absolutePathError: errors.New("absolutePath"),
				})

				So(err, ShouldBeNil)
				So(source, ShouldNotBeNil)

				Convey("calling destroy should not perform any action", func() {
					So(source.destroy(), ShouldBeNil)

					Convey("and the appropriate events should be sent", func() {
						sink.expectEvents(eventSourceCreated, eventSourceDestroyed)
					})
				})

				Convey("calling Files should return an error", func() {
					var results = collectSourceResults(source)

					So(results, ShouldHaveLength, 1)
					So(results[0].File(), ShouldBeNil)
					So(results[0].Error(), ShouldNotBeNil)
					So(results[0].Error().Error(), ShouldEqual, "absolutePath")

					Convey("and the appropriate events should be sent", func() {
						sink.expectEvents(eventSourceCreated, eventSourceStarted, eventSourceResultProduced,
							eventSourceFinished)
					})
				})
			})

		Convey("which contains a number of files and has a Filesystem which performs an action when destroyed", func() {
			source, err = NewSource(SourceConfig{
				ID:      "source",
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

			So(err, ShouldBeNil)
			So(source, ShouldNotBeNil)

			Convey("calling destroy should return the expected value", func() {
				err = source.destroy()

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "destroy")

				Convey("and the appropriate events should be sent", func() {
					sink.expectEvents(eventSourceCreated, eventSourceDestroyed)
				})
			})

			Convey("calling Files", func() {
				var results []Result

				Convey("should return the expected Results", func() {
					results = collectSourceResults(source)

					So(results, ShouldHaveLength, 3)

					expectFilePathsInResults(nil, results, []string{"file", "dir1/file1", "dir2/file2"})

					Convey("and the appropriate events should be sent", func() {
						sink.expectEvents(eventSourceCreated, eventSourceStarted, eventSourceResultProduced,
							eventSourceResultProduced, eventSourceResultProduced, eventSourceFinished)
					})
				})

				Convey("and cancelling should return the expected Results", func(c C) {
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

						c.So(results, ShouldHaveLength, 2)

						expectFilePathsInResults(c, results, []string{"file", "dir1/file1", "dir2/file2"})

						wg.Done()
					})

					wg.Wait()

					Convey("and the appropriate events should be sent", func() {
						sink.expectEvents(eventSourceCreated, eventSourceStarted, eventSourceResultProduced,
							eventSourceResultProduced, eventSourceCancelled, eventSourceFinished)
					})
				})
			})

			Convey("calling ID should return the expected ID", func() {
				So(source.ID(), ShouldEqual, "source")
			})
		})

		Convey("which panics when calling Files", func() {
			source, err = NewSource(SourceConfig{ID: "source"}, &memFilesystem{
				absolutePathError: errors.New("absolutePath"),
				panic:             true,
			})

			So(err, ShouldBeNil)
			So(source, ShouldNotBeNil)

			Convey("calling Files should return an error instead of throwing a panic", func() {
				var results = collectSourceResults(source)

				So(results, ShouldHaveLength, 1)
				So(results[0].File(), ShouldBeNil)
				So(results[0].Error(), ShouldNotBeNil)
				So(results[0].Error().Error(), ShouldEqual, "a fatal error occurred: absolutePath")

				Convey("and the appropriate events should be sent", func() {
					sink.expectEvents(eventSourceCreated, eventSourceStarted, eventSourceResultProduced,
						eventSourceFinished)
				})
			})
		})
	})
}

func TestSourceEvents(t *testing.T) {
	var id = "source"

	Convey("When calling sourceEventCancelled", t, func() {
		validateSourceEvent(sourceEventCancelled(id), componentSource, eventTypeCancelled, id)
	})

	Convey("When calling sourceEventCreated", t, func() {
		validateSourceEvent(sourceEventCreated(id), componentSource, eventTypeCreated, id)
	})

	Convey("When calling sourceEventDestroyed", t, func() {
		validateSourceEvent(sourceEventDestroyed(id), componentSource, eventTypeDestroyed, id)
	})

	Convey("When calling sourceEventFinished", t, func() {
		validateSourceEvent(sourceEventFinished(id), componentSource, eventTypeFinished, id)
	})

	Convey("When calling sourceEventResultProduced", t, func() {
		var res = &result{
			err: errors.New("result error"),
			file: &file{
				fileInfo: &nilFileInfo{
					name: "name",
				},
				path: newFilePath(nil, "name", "/"),
			},
		}

		validateSourceEvent(sourceEventResultProduced(id, res), componentSource, eventTypeResultProduced, id)
	})

	Convey("When calling sourceEventStarted", t, func() {
		validateSourceEvent(sourceEventStarted(id), componentSource, eventTypeStarted, id)
	})
}
