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
	Convey("When creating a Filter", t, func() {
		Convey("which uses a FileEvaluator that discards some files and returns an error on destroy", func() {
			var err error
			var filter Filter
			var source Source

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
			})

			Convey("calling Files", func() {
				Convey("should return the expected Results", func() {
					var results = collectSourceResults(filter)

					expectFilePathsInResults(nil, results, []string{"file1.keep", "file2.keep", "file3.keep",
						"file4.keep"})
				})

				Convey("and cancelling should return the expected Results", func(c C) {
					var cancel CancelFunc
					var in <-chan Result
					var results = make([]Result, 0)
					var wg sync.WaitGroup

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
				})
			})

			Convey("calling ID should return the expected ID", func() {
				So(filter.ID(), ShouldEqual, "filter")
			})
		})

		Convey("which uses a FileEvaluator that panics when calling ShouldKeep", func() {
			var err error
			var filter Filter
			var source Source

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
			})
		})
	})
}

func TestNewFilter(t *testing.T) {
	Convey("When calling NewFilter", t, func() {
		Convey("with a nil Source array", func() {
			var err error
			var filter Filter

			filter, err = NewFilter(FilterConfig{ID: "filter"}, nil, nil)

			Convey("it should return an error", func() {
				So(filter, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errSourceNone), ShouldBeTrue)
			})
		})

		Convey("with an empty Source array", func() {
			var err error
			var filter Filter

			filter, err = NewFilter(FilterConfig{ID: "filter"}, []Source{}, nil)

			Convey("it should return an error", func() {
				So(filter, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errSourceNone), ShouldBeTrue)
			})
		})

		Convey("with a nil FileEvaluator", func() {
			var err error
			var filter Filter
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
			})

			Convey("calling Files should return all files", func() {
				var results = collectSourceResults(filter)

				So(results, ShouldHaveLength, 2)

				expectFilePathsInResults(nil, results, []string{"file1", "file2"})
			})
		})
	})
}

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
