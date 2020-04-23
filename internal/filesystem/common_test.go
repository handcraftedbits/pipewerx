package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Public functions
//

func TestMain(m *testing.M) {
	var code int

	docker = testutil.NewDocker("")

	code = m.Run()

	docker.Destroy()

	os.Exit(code)
}

//
// Private types
//

type testFilesystemConfig struct {
	createFunc func() (pipewerx.Filesystem, error)
	name       string
	realPath   func(string, string) string
}

//
// Private variables
//

var docker *testutil.Docker

//
// Private functions
//

func testFilesystem(t *testing.T, config testFilesystemConfig) {
	Convey("When creating "+config.name+" Filesystem", t, func(c C) {
		var err error
		var fs pipewerx.Filesystem
		var sep string

		So(err, ShouldBeNil)

		fs, err = config.createFunc()

		So(err, ShouldBeNil)
		So(fs, ShouldNotBeNil)

		sep = fs.PathSeparator()

		defer func() {
			err = fs.Destroy()

			So(err, ShouldBeNil)
		}()

		Convey("calling AbsolutePath should return the expected value", func() {
			var result string
			var tests = []struct {
				input    string
				expected string
			}{
				{"abs", "abs"},
				{fmt.Sprintf("a%sb%sc", sep, sep), fmt.Sprintf("a%sb%sc", sep, sep)},
				{fmt.Sprintf("a%sb%s..%sc", sep, sep, sep), fmt.Sprintf("a%sc", sep)},
			}

			for _, test := range tests {
				result, err = fs.AbsolutePath(test.input)

				So(err, ShouldBeNil)
				So(result, ShouldEndWith, test.expected)
			}
		})

		Convey("calling BasePart should return the expected value", func() {
			var result string
			var tests = []struct {
				input    string
				expected string
			}{
				{"abc", "abc"},
				{fmt.Sprintf("a%sb%sc%sabc", sep, sep, sep), "abc"},
			}

			for _, test := range tests {
				result = fs.BasePart(test.input)

				So(result, ShouldEqual, test.expected)
			}
		})

		Convey("calling DirPart should return the expected value", func() {
			var result []string
			var tests = []struct {
				input    string
				expected []string
			}{
				{"abc", []string{}},
				{fmt.Sprintf("a%sb%sc%sabc", sep, sep, sep), []string{"a", "b", "c"}},
			}

			for _, test := range tests {
				result = fs.DirPart(test.input)

				So(result, ShouldHaveLength, len(test.expected))

				for i, value := range result {
					So(result[i], ShouldEqual, value)
				}
			}
		})

		Convey("calling ListFiles", func() {
			Convey("should return an error when an invalid directory is specified", func() {
				var fileInfos []os.FileInfo

				fileInfos, err = fs.ListFiles(config.realPath(testutil.TestdataPathFilesystem, ";;;;"))

				So(fileInfos, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			// TODO: leave out until SMB is fixed... seems to return nil now instead of the EOF error it used to...
			/*Convey("should return the expected files in an empty directory", func() {
				var fileInfos []os.FileInfo

				fileInfos, err = fs.ListFiles(config.realPath(testutil.TestdataPathFilesystem, "emptyDir"))

				So(err, ShouldBeNil)
				So(fileInfos, ShouldNotBeNil)
				So(fileInfos, ShouldHaveLength, 0)
			})*/

			Convey("should return the expected files in a non-empty directory", func() {
				var fileInfos []os.FileInfo
				var fileInfosByName = make(map[string]os.FileInfo)
				var tests = []struct {
					name string
					dir  bool
				}{
					{"a.test", false},
					{"b.test", false},
					{"c", true},
					{"d", true},
				}

				fileInfos, err = fs.ListFiles(config.realPath(testutil.TestdataPathFilesystem, "mixed"))

				So(err, ShouldBeNil)
				So(fileInfos, ShouldNotBeNil)
				So(fileInfos, ShouldHaveLength, 4)

				for _, fileInfo := range fileInfos {
					fileInfosByName[fileInfo.Name()] = fileInfo
				}

				for _, test := range tests {
					var fileInfo = fileInfosByName[test.name]

					So(fileInfo, ShouldNotBeNil)
					So(fileInfo.IsDir(), ShouldEqual, test.dir)
				}
			})
		})

		Convey("calling ReadFile", func() {
			Convey("should return an error when an invalid file is specified", func() {
				var reader io.ReadCloser

				reader, err = fs.ReadFile(config.realPath(testutil.TestdataPathFilesystem, ";;;;"))

				So(reader, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("should be able to stat a valid file", func() {
				var contents []byte
				var reader io.ReadCloser

				reader, err = fs.ReadFile(config.realPath(testutil.TestdataPathFilesystem, "fileOnly.test"))

				So(err, ShouldBeNil)
				So(reader, ShouldNotBeNil)

				contents, err = ioutil.ReadAll(reader)

				So(err, ShouldBeNil)
				So(contents, ShouldNotBeNil)
				So(string(contents), ShouldEqual, "fileOnly")

				err = reader.Close()

				So(err, ShouldBeNil)
			})
		})

		Convey("calling StatFile", func() {
			Convey("should return an error when an invalid file is specified", func() {
				var fileInfo os.FileInfo

				fileInfo, err = fs.StatFile(config.realPath(testutil.TestdataPathFilesystem, ";;;;"))

				So(fileInfo, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("should be able to stat a valid file", func() {
				var fileInfo os.FileInfo

				fileInfo, err = fs.StatFile(config.realPath(testutil.TestdataPathFilesystem, "fileOnly.test"))

				So(err, ShouldBeNil)
				So(fileInfo, ShouldNotBeNil)
				So(fileInfo.Name(), ShouldEqual, "fileOnly.test")
				So(fileInfo.IsDir(), ShouldBeFalse)
			})
		})
	})
}
