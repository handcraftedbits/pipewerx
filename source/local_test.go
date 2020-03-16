package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/client"
)

//
// Testcases
//

// localFileProducer tests

func TestNewLocal(t *testing.T) {
	Convey("When creating a new local Source", t, func() {
		Convey("it should return an error when accessing the first result if a malformed root is provided", func() {
			var in <-chan pipewerx.Result
			var results = make([]pipewerx.Result, 0)

			var source = NewLocal(&LocalConfig{
				Name:    "local",
				Recurse: false,
				Root:    "???abc???",
			})

			in, _ = source.Files(pipewerx.NewContext())

			for result := range in {
				results = append(results, result)
			}

			So(results, ShouldHaveLength, 1)
			So(results[0].File(), ShouldBeNil)
			So(results[0].Error(), ShouldNotBeNil)
		})
	})
}

func TestLocalFileProducer_Next(t *testing.T) {
	Convey("When creating a local Source", t, func() {
		testFileProducer(testDataRoot, func(root string, recurse bool) pipewerx.Source {
			return NewLocal(&LocalConfig{
				Name:    "local",
				Recurse: recurse,
				Root:    root,
			})
		}, func(stepper *pathStepper) pipewerx.FileProducer {
			return &localFileProducer{
				stepper: stepper,
			}
		}, func() client.Filesystem {
			return localFSInstance
		})
	})
}
