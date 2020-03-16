package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/client"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Public functions
//

func TestMain(m *testing.M) {
	var code int
	var contents []byte
	var err error

	contents, err = ioutil.ReadFile("testdata/sourceTests.json")

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(contents, &fileProducerTests)

	if err != nil {
		panic(err)
	}

	docker = testutil.NewDocker("")

	code = m.Run()

	docker.Destroy()

	os.Exit(code)
}

//
// Private types
//

type expectedResult struct {
	Contents string `json:"contents"`
	Path     string `json:"path"`
}

type fileProducerTest struct {
	Description   string           `json:"description"`
	PathsToCreate []string         `json:"pathsToCreate"`
	Recurse       bool             `json:"recurse"`
	Results       []expectedResult `json:"results"`
	Root          string           `json:"root"`
}

type fileProducerBuilderFunc func(*pathStepper) pipewerx.FileProducer
type filesystemBuilderFunc func() client.Filesystem
type sourceBuilderFunc func(string, bool) pipewerx.Source

//
// Private variables
//

var (
	docker            *testutil.Docker
	fileProducerTests []fileProducerTest
	testDataRoot      = "testdata/fileProducer/"
)

//
// Private functions
//

func expectResultsForRoot(source pipewerx.Source, fs client.Filesystem, expected []expectedResult) {
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
		// The expectation is that all paths in expectedResult use '/', so for filesystems that have a different
		// separator we have to replace all occurrences.

		expectedMap[strings.ReplaceAll(value.Path, "/", fs.PathSeparator())] = value.Contents
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

func testFileProducer(fsPrefix string, sourceBuilder sourceBuilderFunc, fileProducer fileProducerBuilderFunc,
	filesystem filesystemBuilderFunc) {
	Convey("it should produce correct Results", func() {
		for _, test := range fileProducerTests {
			Convey("when "+test.Description, func() {
				var source pipewerx.Source

				for _, path := range test.PathsToCreate {
					var err = os.MkdirAll(testDataRoot+test.Root+"/"+path, 0775)

					So(err, ShouldBeNil)
				}

				source = sourceBuilder(fsPrefix+test.Root, test.Recurse)

				expectResultsForRoot(source, filesystem(), test.Results)
			})
		}
	})

	Convey("it should return an error", func() {
		Convey("when an error occurs while stepping through files", func() {
			var err error
			var file pipewerx.File
			var listFilesError = errors.New("listFiles")
			var stepper *pathStepper

			stepper, err = newPathStepper(filesystem(), fsPrefix+"singleLevelSubdirs", true)

			So(err, ShouldBeNil)
			So(stepper, ShouldNotBeNil)

			stepper.fs = newErrorFilesystem(stepper.fs, nil, listFilesError)

			file, err = fileProducer(stepper).Next()

			So(file, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "listFiles")
		})
	})
}
