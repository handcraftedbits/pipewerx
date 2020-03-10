package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"golang.handcraftedbits.com/pipewerx"
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
	Convey("When creating a new local Source", t, func() {
		Convey("it should produce correct Results", func() {
			Convey("when an empty directory is used as the root", func() {
				var tempDirErr = os.MkdirAll("testdata/local/emptyDir", 0775)

				So(tempDirErr, ShouldBeNil)

				var source = NewLocal(&LocalConfig{
					Name:    "local",
					Recurse: true,
					Root:    "testdata/local/emptyDir",
				})

				expectResultsForRoot(source, []expectedResult{})
			})

			Convey("when a directory containing multiple empty directories is used as the root", func() {
				var err = os.MkdirAll("testdata/local/multipleEmptyDirs", 0775)

				So(err, ShouldBeNil)

				err = os.MkdirAll("testdata/local/multipleEmptyDirs/a", 0775)

				So(err, ShouldBeNil)

				err = os.MkdirAll("testdata/local/multipleEmptyDirs/b/c", 0775)

				So(err, ShouldBeNil)

				err = os.MkdirAll("testdata/local/multipleEmptyDirs/d/e/f", 0775)

				So(err, ShouldBeNil)

				var source = NewLocal(&LocalConfig{
					Name:    "local",
					Recurse: true,
					Root:    "testdata/local/multipleEmptyDirs",
				})

				expectResultsForRoot(source, []expectedResult{})
			})

			Convey("when a single file is used as the root", func() {
				var source = NewLocal(&LocalConfig{
					Name:    "local",
					Recurse: true,
					Root:    "testdata/local/fileOnly.test",
				})

				expectResultsForRoot(source, []expectedResult{
					{path: "fileOnly.test", contents: "fileOnly"},
				})
			})

			Convey("when there are no subdirectories", func() {
				var source = NewLocal(&LocalConfig{
					Name:    "local",
					Recurse: true,
					Root:    "testdata/local/filesOnly",
				})

				expectResultsForRoot(source, []expectedResult{
					{path: "a.test", contents: "a"},
					{path: "b.test", contents: "b"},
					{path: "c.test", contents: "c"},
				})
			})

			Convey("when there is a single level of subdirectories", func() {
				Convey("and recursion is enabled", func() {
					var source = NewLocal(&LocalConfig{
						Name:    "local",
						Recurse: true,
						Root:    "testdata/local/singleLevelSubdirs",
					})

					expectResultsForRoot(source, []expectedResult{
						{path: "a/a.test", contents: "a"},
						{path: "b/b.test", contents: "b"},
						{path: "c/c.test", contents: "c"},
					})
				})

				Convey("and recursion is disabled", func() {
					var source = NewLocal(&LocalConfig{
						Name:    "local",
						Recurse: false,
						Root:    "testdata/local/singleLevelSubdirs",
					})

					expectResultsForRoot(source, []expectedResult{})
				})
			})

			Convey("when there are multiple levels of subdirectories", func() {
				Convey("and recursion is enabled", func() {
					var source = NewLocal(&LocalConfig{
						Name:    "local",
						Recurse: true,
						Root:    "testdata/local/multiLevelSubdirs",
					})

					expectResultsForRoot(source, []expectedResult{
						{path: "a/a.test", contents: "a"},
						{path: "b/c/c.test", contents: "c"},
						{path: "d/e/f/f.test", contents: "f"},
					})
				})

				Convey("and recursion is disabled", func() {
					var source = NewLocal(&LocalConfig{
						Name:    "local",
						Recurse: false,
						Root:    "testdata/local/multiLevelSubdirs",
					})

					expectResultsForRoot(source, []expectedResult{})
				})
			})

			Convey("when there is a mixture of subdirectory depths", func() {
				Convey("and recursion is enabled", func() {
					var source = NewLocal(&LocalConfig{
						Name:    "local",
						Recurse: true,
						Root:    "testdata/local/mixed",
					})

					expectResultsForRoot(source, []expectedResult{
						{path: "a.test", contents: "a"},
						{path: "b.test", contents: "b"},
						{path: "c/c.test", contents: "c"},
						{path: "d/e/f/f.test", contents: "f"},
					})
				})

				Convey("and recursion is disabled", func() {
					var source = NewLocal(&LocalConfig{
						Name:    "local",
						Recurse: false,
						Root:    "testdata/local/mixed",
					})

					expectResultsForRoot(source, []expectedResult{
						{path: "a.test", contents: "a"},
						{path: "b.test", contents: "b"},
					})
				})
			})
		})

		Convey("it should return an error", func() {
			Convey("when an error occurs while stepping through files", func() {
				var err error
				var file pipewerx.File
				var listFilesError = errors.New("listFiles")
				var producer *localFileProducer
				var stepper *pathStepper

				stepper, err = newPathStepper(localFSInstance, "testdata/local/singleLevelSubdirs", true)

				So(err, ShouldBeNil)
				So(stepper, ShouldNotBeNil)

				stepper.fs = newErrorFilesystem(nil, listFilesError)

				producer = &localFileProducer{
					stepper: stepper,
				}

				file, err = producer.Next()

				So(file, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "listFiles")
			})
		})
	})
}

//
// Private types
//

type expectedResult struct {
	contents string
	path     string
}

//
// Private functions
//

func expectResultsForRoot(source pipewerx.Source, expected []expectedResult) {
	var expectedMap = make(map[string]string)
	var in <-chan pipewerx.Result
	var results = make([]pipewerx.Result, 0)

	in, _ = source.Files(pipewerx.NewContext())

	for result := range in {
		So(result.Error(), ShouldBeNil)
		So(result.File(), ShouldNotBeNil)

		results = append(results, result)
	}

	for _, value := range expected {
		expectedMap[value.path] = value.contents
	}

	So(results, ShouldHaveLength, len(expected))

	for _, result := range results {
		var contents []byte
		var err error
		var reader io.ReadCloser

		So(result.Error(), ShouldBeNil)
		So(result.File(), ShouldNotBeNil)
		So(expectedMap, ShouldContainKey, result.File().Path().String())

		reader, err = result.File().Reader()

		So(err, ShouldBeNil)
		So(reader, ShouldNotBeNil)

		contents, err = ioutil.ReadAll(reader)

		So(err, ShouldBeNil)
		So(expectedMap[result.File().Path().String()], ShouldEqual, string(contents))

		err = reader.Close()

		So(err, ShouldBeNil)
	}
}
