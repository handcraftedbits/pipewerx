package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Private types
//

type testSourceConfig struct {
	createFunc    func(id, root string, recurse bool) (pipewerx.Source, error)
	name          string
	pathSeparator string
	realPath      func(string, string) string
}

//
// Private functions
//

func collectSourceResults(source pipewerx.Source) []pipewerx.Result {
	var in <-chan pipewerx.Result
	var results = make([]pipewerx.Result, 0)

	in, _ = source.Files(pipewerx.NewContext(pipewerx.ContextConfig{}))

	for result := range in {
		results = append(results, result)
	}

	return results
}

func expectFilesInResults(results []pipewerx.Result, separator string, paths []string, contents []string) {
	var contentMap = make(map[string]string)
	var pathMap = make(map[string]bool)

	for i, path := range paths {
		// For simplicity's sake we're using '/' as a path separator, but that might not be true for all filesystems.
		// To address that, replace all '/' with the correct path separator before adding to the map.

		path = strings.ReplaceAll(path, "/", separator)

		pathMap[path] = true

		if contents != nil {
			contentMap[path] = contents[i]
		}
	}

	for _, result := range results {
		So(result.Error(), ShouldBeNil)
		So(pathMap, ShouldContainKey, result.File().Path().String())

		if contents != nil {
			var err error
			var fileContents []byte
			var reader io.ReadCloser

			So(contentMap, ShouldContainKey, result.File().Path().String())

			reader, err = result.File().Reader()

			So(err, ShouldBeNil)
			So(reader, ShouldNotBeNil)

			fileContents, err = ioutil.ReadAll(reader)

			So(err, ShouldBeNil)
			So(fileContents, ShouldNotBeNil)

			err = reader.Close()

			So(err, ShouldBeNil)
			So(result.File().Size(), ShouldEqual, len(fileContents))
			So(string(fileContents), ShouldEqual, contentMap[result.File().Path().String()])
		}
	}
}

func mustCreateSource(config testSourceConfig, id, root string, recurse bool) pipewerx.Source {
	var err error
	var source pipewerx.Source

	source, err = config.createFunc(id, root, recurse)

	So(err, ShouldBeNil)
	So(source, ShouldNotBeNil)

	return source
}

func newSMBConfig(port int) SMBConfig {
	return SMBConfig{
		Domain:   testutil.ConstSMBDomain,
		Host:     "localhost",
		Password: testutil.ConstSMBPassword,
		Port:     port,
		Share:    testutil.ConstSMBShare,
		Username: testutil.ConstSMBUser,
	}
}

func testSource(t *testing.T, config testSourceConfig) {
	var sourceName = "source"

	Convey("Creating "+config.name+" Source should fail when an invalid ID is provided", t, func() {
		var err error
		var ids = []string{"", " ", ".", "a ", " a", "a.", ".a", "a..b", "a-b", "?"}
		var source pipewerx.Source

		for _, id := range ids {
			source, err = config.createFunc(id, "", false)

			So(source, ShouldBeNil)
			So(err, ShouldNotBeNil)
		}
	})

	Convey("When creating "+config.name+" Source", t, func() {
		var results []pipewerx.Result
		var root string

		Convey("calling Files should return the expected values", func() {
			Convey("when an empty directory is used as the root", func() {
				root = config.realPath(testutil.TestdataPathFilesystem, "emptyDir")

				Convey("and recursion is enabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

					expectFilesInResults(results, config.pathSeparator, []string{}, nil)
				})

				Convey("and recursion is disabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

					expectFilesInResults(results, config.pathSeparator, []string{}, nil)
				})
			})

			Convey("when a directory containing multiple empty directories is used as the root", func() {
				root = config.realPath(testutil.TestdataPathFilesystem, "multipleEmptyDirs")

				Convey("and recursion is enabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

					expectFilesInResults(results, config.pathSeparator, []string{}, nil)
				})

				Convey("and recursion is disabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

					expectFilesInResults(results, config.pathSeparator, []string{}, nil)
				})
			})

			Convey("when a single file is used as the root", func() {
				root = config.realPath(testutil.TestdataPathFilesystem, "fileOnly.test")

				Convey("and recursion is enabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

					expectFilesInResults(results, config.pathSeparator, []string{"fileOnly.test"}, []string{"fileOnly"})
				})

				Convey("and recursion is disabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

					expectFilesInResults(results, config.pathSeparator, []string{"fileOnly.test"}, []string{"fileOnly"})
				})
			})

			Convey("when there are no subdirectories", func() {
				root = config.realPath(testutil.TestdataPathFilesystem, "filesOnly")

				Convey("and recursion is enabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

					expectFilesInResults(results, config.pathSeparator, []string{"a.test", "b.test", "c.test"},
						[]string{"a", "b", "c"})
				})

				Convey("and recursion is disabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

					expectFilesInResults(results, config.pathSeparator, []string{"a.test", "b.test", "c.test"},
						[]string{"a", "b", "c"})
				})
			})

			Convey("when there is a single level of subdirectories", func() {
				root = config.realPath(testutil.TestdataPathFilesystem, "singleLevelSubdirs")

				Convey("and recursion is enabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

					expectFilesInResults(results, config.pathSeparator, []string{"a/a.test", "b/b.test", "c/c.test"},
						[]string{"a", "b", "c"})
				})

				Convey("and recursion is disabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

					expectFilesInResults(results, config.pathSeparator, []string{}, nil)
				})
			})

			Convey("when there are multiple levels of subdirectories", func() {
				root = config.realPath(testutil.TestdataPathFilesystem, "multiLevelSubdirs")

				Convey("and recursion is enabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

					expectFilesInResults(results, config.pathSeparator, []string{"a/a.test", "b/c/c.test",
						"d/e/f/f.test"}, []string{"a", "c", "f"})
				})

				Convey("and recursion is disabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

					expectFilesInResults(results, config.pathSeparator, []string{}, nil)
				})
			})

			Convey("when there is a mixture of subdirectory depths", func() {
				root = config.realPath(testutil.TestdataPathFilesystem, "mixed")

				Convey("and recursion is enabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

					expectFilesInResults(results, config.pathSeparator, []string{"a.test", "b.test", "c/c.test",
						"d/e/f/f.test"}, []string{"a", "b", "c", "f"})
				})

				Convey("and recursion is disabled", func() {
					results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

					expectFilesInResults(results, config.pathSeparator, []string{"a.test", "b.test"},
						[]string{"a", "b"})
				})
			})
		})
	})
}
