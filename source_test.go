package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// mergedSource tests

func TestNewMergedSource(t *testing.T) {
	Convey("When creating a merged Source", t, func() {
		Convey("providing a nil array of Sources should create an empty Source", func() {
			var in <-chan Result
			var results []Result
			var source = newMergedSource(nil)

			in, _ = source.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 0)
		})

		Convey("providing an array of nil Sources should create an empty Source", func() {
			var in <-chan Result
			var results []Result
			var source = newMergedSource(nil, nil, nil)

			in, _ = source.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 0)
		})

		Convey("all nil Sources are removed", func() {
			var in <-chan Result
			var results []Result
			var source = newMergedSource(nil, NewSource("name", newSimpleResultProducer()), nil)

			in, _ = source.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 3)
		})

		Convey("any duplicate Source references must be removed", func() {
			var in <-chan Result
			var mergedSource Source
			var results []Result
			var source = NewSource("name", newSimpleResultProducer())

			mergedSource = newMergedSource(source, source, source)

			in, _ = mergedSource.Files(newContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 3)
		})
	})
}
