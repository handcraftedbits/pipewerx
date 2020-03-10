package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// mergedSource tests

func TestMergedSource_Items(t *testing.T) {
	Convey("When creating a merged Source", t, func() {
		Convey("it should produce an error Result when an initialization error occurs", func() {
			var in <-chan Result
			var index = -1
			var results = make([]Result, 0)
			var source = newMergedSource([]Source{newSimpleSource("simple", "/a/b", 5),
				newErrorSource("init", "/a/b", 5, true, false, false)})

			in, _ = source.Files(NewContext())

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
			So(results[index].Error().Error(), ShouldEqual, "init")
		})

		Convey("it should produce an error Result when a destruction error occurs", func() {
			var in <-chan Result
			var index = -1
			var results = make([]Result, 0)
			var source = newMergedSource([]Source{newSimpleSource("simple", "/a/b", 5),
				newErrorSource("destroy", "/a/b", 5, false, true, false)})

			in, _ = source.Files(NewContext())

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

		Convey("it should be able to produce all Results when no errors occur", func() {
			var in <-chan Result
			var results = make(map[string]bool)
			var source = newSimpleMergedSource(3)

			in, _ = source.Files(NewContext())

			for result := range in {
				So(result.Error(), ShouldBeNil)
				So(result.File(), ShouldNotBeNil)

				results[result.File().Path().String()] = true
			}

			// Use the map as a hash set.  Since all the test paths are unique, we should end up with six keys.

			So(results, ShouldHaveLength, 6)
		})

		Convey("it should stop when the Source is terminated", func() {
			var cancel func()
			var in <-chan Result
			var results = make([]Result, 0)
			var source = newSimpleMergedSource(5)

			in, cancel = source.Files(NewContext())

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

func TestMergedSource_Name(t *testing.T) {
	Convey("When creating a merged Source", t, func() {
		Convey("it should have the correct name when a nil Source array is provided", func() {
			var source = newMergedSource(nil)

			So(source.Name(), ShouldEqual, "<empty source>")
		})

		Convey("it should have the correct name when a valid Source array is provided", func() {
			var source = newSimpleMergedSource(3)

			So(source.Name(), ShouldEqual, "<merged source>")
		})
	})
}

func TestNewMergedSource(t *testing.T) {
	Convey("When creating a merged Source", t, func() {
		Convey("providing a nil array of Sources should create an empty Source", func() {
			var in <-chan Result
			var results []Result
			var source = newMergedSource(nil)

			in, _ = source.Files(NewContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 0)
		})

		Convey("providing an array of nil Sources should create an empty Source", func() {
			var in <-chan Result
			var results []Result
			var source = newMergedSource([]Source{nil, nil, nil})

			in, _ = source.Files(NewContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 0)
		})

		Convey("all nil Sources are removed", func() {
			var in <-chan Result
			var results []Result
			var source = newMergedSource([]Source{nil, newSimpleSource("name", "/a/b", 3), nil})

			in, _ = source.Files(NewContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 3)
		})

		Convey("any duplicate Source references must be removed", func() {
			var in <-chan Result
			var mergedSource Source
			var results []Result
			var source = newSimpleSource("name", "/a/b", 3)

			mergedSource = newMergedSource([]Source{source, source, source})

			in, _ = mergedSource.Files(NewContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 3)
		})
	})
}

// Source tests

func TestNewSource(t *testing.T) {
	Convey("When creating a Source with a nil FileProducer function", t, func() {
		var source = NewSource("source", nil)

		Convey("it should produce no Results", func() {
			var in <-chan Result
			var results []Result

			in, _ = source.Files(NewContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 0)
		})
	})

	Convey("When creating a Source with a FileProducer function that returns nil", t, func() {
		var source = NewSource("source", func(Context) (FileProducer, error) {
			return nil, nil
		})

		Convey("it should produce a single error Result", func() {
			var in <-chan Result
			var results = make([]Result, 0)

			in, _ = source.Files(NewContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 1)
			So(results[0].File(), ShouldBeNil)
			So(results[0].Error(), ShouldNotBeNil)
		})
	})
}

func TestSource_Items(t *testing.T) {
	Convey("When creating a Source", t, func() {
		Convey("it should produce an error Result when an initialization error occurs", func() {
			var in <-chan Result
			var results = make([]Result, 0)
			var source = newErrorSource("init", "/a/b", 5, true, false, false)

			in, _ = source.Files(NewContext())

			for result := range in {
				results = append(results, result)
			}

			So(len(results), ShouldEqual, 1)
			So(results[0].File(), ShouldBeNil)
			So(results[0].Error(), ShouldNotBeNil)
			So(results[0].Error().Error(), ShouldEqual, "init")
		})

		Convey("it should produce an error Result when a destruction error occurs", func() {
			var in <-chan Result
			var index = -1
			var results = make([]Result, 0)
			var source = newErrorSource("destroy", "/a/b", 5, false, true, false)

			in, _ = source.Files(NewContext())

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

		Convey("it should be able to produce all Results when no errors occur", func() {
			var in <-chan Result
			var results = make(map[string]bool)
			var source = newSimpleSource("simple", "/a/b", 3)

			in, _ = source.Files(NewContext())

			for result := range in {
				So(result.Error(), ShouldBeNil)
				So(result.File(), ShouldNotBeNil)

				results[result.File().Path().String()] = true
			}

			// Use the map as a hash set.  Since all the test paths are unique, we should end up with three keys.

			So(results, ShouldHaveLength, 3)
		})

		Convey("it should stop when the Source is terminated", func() {
			var cancel func()
			var in <-chan Result
			var results = make([]Result, 0)
			var source = newSimpleSource("simple", "/a/b", 10)

			in, cancel = source.Files(NewContext())

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

func TestSource_Name(t *testing.T) {
	Convey("When creating a Source", t, func() {
		var source = NewSource("name", nil)

		Convey("it should have the correct name", func() {
			So(source.Name(), ShouldEqual, "name")
		})
	})
}

//
// Private functions
//

func newErrorSource(name, pathPrefix string, size int, duringCreate, duringDestroy, duringProduce bool) Source {
	var destroyError error
	var produceError error

	if duringCreate {
		return NewSource(name, func(Context) (FileProducer, error) {
			return nil, errors.New(name)
		})
	}

	if duringDestroy {
		destroyError = errors.New(name)
	}

	if duringProduce {
		produceError = errors.New(name)
	}

	return NewSource(name, func(Context) (FileProducer, error) {
		return newTestFileProducer(pathPrefix, size, destroyError, produceError), nil
	})
}

func newSimpleMergedSource(size int) Source {
	return newMergedSource([]Source{newSimpleSource("first", "/first", size),
		newSimpleSource("second", "/second", size)})
}

func newSimpleSource(name, pathPrefix string, numFiles int) Source {
	return NewSource(name, func(Context) (FileProducer, error) {
		return newSimpleFileProducer(pathPrefix, numFiles), nil
	})
}
