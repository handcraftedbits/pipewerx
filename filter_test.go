package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// Filter tests

func TestFilter_Files(t *testing.T) {
	Convey("When creating a Filter", t, func() {
		Convey("it should return an error Result when an initialization error occurs", func() {
			var filter = newErrorFilter("init", "/a/b", 3, 10, true, false, false)
			var in <-chan Result
			var results = make([]Result, 0)

			in, _ = filter.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(len(results), ShouldEqual, 1)
			So(results[0].File(), ShouldBeNil)
			So(results[0].Error(), ShouldNotBeNil)
			So(results[0].Error().Error(), ShouldEqual, "init")
		})

		Convey("it should return an error Result when a Source destruction error occurs", func() {
			var filter = newErrorFilter("destroy", "/a/b", 3, 10, false, true, false)
			var in <-chan Result
			var index = -1
			var results = make([]Result, 0)

			in, _ = filter.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(len(results), ShouldBeGreaterThanOrEqualTo, 1)

			for i, result := range results {
				if result.Error() != nil {
					index = i

					break
				}
			}

			So(index, ShouldNotEqual, -1)
			So(results[index].File(), ShouldBeNil)
			So(results[index].Error(), ShouldNotBeNil)
			So(results[index].Error().Error(), ShouldEqual, "destroy")
		})

		Convey("it should be able to return all Results when no errors occur", func() {
			var filter = newSimpleFilter("filter", "/a/b", 3)
			var in <-chan Result
			var results = make(map[string]bool)

			in, _ = filter.Files(newContext())

			for result := range in {
				So(result.Error(), ShouldBeNil)
				So(result.File(), ShouldNotBeNil)

				results[result.File().Path().String()] = true
			}

			// Use the map as a hash set.  Since all the test paths are unique, we should end up with three keys.

			So(results, ShouldHaveLength, 3)
		})

		Convey("it should return no Results when a FileEvaluator that discards all items is used", func() {
			var filter = NewFilter("filter", func(Context) (FileEvaluator, error) {
				return &discardingFileEvaluator{}, nil
			}, []Source{newSimpleSource("source", "/a/b", 5)})
			var in <-chan Result
			var results = make([]Result, 0)

			in, _ = filter.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldBeEmpty)
		})

		Convey("it should return an error Result when the underlying Source returns an error Result", func() {
			var filter = NewFilter("filter", nil, []Source{newErrorSource("produce", "/a/b", 3, false, false, true)})
			var in <-chan Result
			var index = -1
			var results = make([]Result, 0)

			in, _ = filter.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(len(results), ShouldBeGreaterThanOrEqualTo, 1)

			for i, result := range results {
				if result.Error() != nil {
					index = i

					break
				}
			}

			So(index, ShouldNotEqual, -1)
			So(results[index].File(), ShouldBeNil)
			So(results[index].Error(), ShouldNotBeNil)
			So(results[index].Error().Error(), ShouldEqual, "produce")
		})

		Convey("it should return an error Result when the FileEvaluator returns an error", func() {
			var filter = newErrorFilter("keep", "/a/b", 10, 5, false, false, true)
			var in <-chan Result
			var index = -1
			var results = make([]Result, 0)

			in, _ = filter.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(len(results), ShouldBeGreaterThanOrEqualTo, 1)

			for i, result := range results {
				if result.Error() != nil {
					index = i

					break
				}
			}

			So(index, ShouldNotEqual, -1)
			So(results[index].File(), ShouldBeNil)
			So(results[index].Error(), ShouldNotBeNil)
			So(results[index].Error().Error(), ShouldEqual, "keep")
		})

		Convey("it should stop when the Filter is terminated", func() {
			var cancel func()
			var filter = newSimpleFilter("filter", "/a/b", 10)
			var in <-chan Result
			var results = make([]Result, 0)

			in, cancel = filter.Files(newContext())

			results = append(results, <-in)
			results = append(results, <-in)

			cancel()

			for result := range in {
				results = append(results, result)
			}

			So(len(results), ShouldBeGreaterThanOrEqualTo, 2)
			So(len(results), ShouldBeLessThan, 10)
		})
	})
}

func TestFilter_Name(t *testing.T) {
	Convey("When creating a Filter", t, func() {
		var filter = NewFilter("name", nil, nil)

		Convey("it should have the correct name", func() {
			So(filter.Name(), ShouldEqual, "name")
		})
	})
}

func TestNewFilter(t *testing.T) {
	Convey("When creating a Filter with a nil FileEvaluator function", t, func() {
		var filter = newSimpleFilter("filter", "/a/b", 3)

		Convey("it should not filter any Results", func() {
			var in <-chan Result
			var results []Result

			in, _ = filter.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 3)
		})
	})

	Convey("When creating a Filter with a FileEvaluator function that returns nil", t, func() {
		var filter = NewFilter("filter", func(Context) (FileEvaluator, error) {
			return nil, nil
		}, []Source{newSimpleSource("source", "/a/b", 3)})

		Convey("it should return a single error Result", func() {
			var in <-chan Result
			var results = make([]Result, 0)

			in, _ = filter.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 1)
			So(results[0].File(), ShouldBeNil)
			So(results[0].Error(), ShouldNotBeNil)
		})
	})
}

//
// Private functions
//

func newErrorFilter(name, pathPrefix string, numFiles, maxEvaluations int, duringCreate, duringDestroy,
	duringKeep bool) Filter {
	var destroyError error
	var keepError error
	var source = newSimpleSource(name, pathPrefix, numFiles)

	if duringDestroy {
		destroyError = errors.New(name)
	}

	if duringKeep {
		keepError = errors.New(name)
	}

	if duringCreate {
		return NewFilter(name, func(Context) (FileEvaluator, error) {
			return nil, errors.New(name)
		}, []Source{source})
	}

	return NewFilter(name, func(Context) (FileEvaluator, error) {
		return newTestFileEvaluator(maxEvaluations, destroyError, keepError), nil
	}, []Source{source})
}

func newSimpleFilter(name, pathPrefix string, numFiles int) Filter {
	return NewFilter(name, nil, []Source{newSimpleSource("source", pathPrefix, numFiles)})
}
