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
		Convey("which contains a number of Sources, each of which have a Filesystem that does nothing when destroyed",
			func() {
				var err error
				var merged Source
				var source [2]Source

				source[0], err = NewSource(SourceConfig{}, &memFilesystem{
					root: &memFilesystemNode{
						children: map[string]*memFilesystemNode{
							"file1": {},
						},
					},
				})

				So(err, ShouldBeNil)
				So(source[0], ShouldNotBeNil)

				source[1], err = NewSource(SourceConfig{}, &memFilesystem{
					root: &memFilesystemNode{
						children: map[string]*memFilesystemNode{
							"file2": {},
						},
					},
				})

				So(err, ShouldBeNil)
				So(source[1], ShouldNotBeNil)

				merged, err = newMergedSource([]Source{source[0], source[1]})

				So(err, ShouldBeNil)
				So(merged, ShouldNotBeNil)

				Convey("calling destroy should return nil", func() {
					So(merged.destroy(), ShouldEqual, nil)
				})
			})

		Convey("which contains a number of Sources, most of which have a Filesystem that performs an action when "+
			"destroyed", func() {
			var err error
			var expected = []string{"file01", "file02", "file03", "file04", "file05", "file06", "file07", "file08",
				"file09", "file10", "file11", "file12", "file13", "file14", "file15"}
			var merged Source
			var source [3]Source

			source[0], err = NewSource(SourceConfig{
				Name: "source1",
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
				Name: "source2",
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
				Name: "source3",
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

			merged, err = newMergedSource([]Source{source[0], source[1], source[2]})

			So(err, ShouldBeNil)
			So(merged, ShouldNotBeNil)

			Convey("calling Destroy should return the expected value", func() {
				var multiErr MultiError

				err = merged.destroy()

				So(err, ShouldNotBeNil)
				So(err, ShouldImplement, (*MultiError)(nil))

				multiErr = err.(MultiError)

				So(multiErr.Causes(), ShouldHaveLength, 2)
				So(multiErr.Causes()[0].Error(), ShouldEqual, "source1")
				So(multiErr.Causes()[1].Error(), ShouldEqual, "source2")
				So(multiErr.Error(), ShouldEqual, sourceMergedDestroyMessage)
			})

			Convey("calling Files", func() {
				Convey("should return the expected Results", func() {
					var results = collectSourceResults(merged)

					So(results, ShouldHaveLength, 15)

					expectFilePathsInResults(nil, results, expected)
				})

				Convey("and cancelling should return the expected Results", func(c C) {
					var cancel CancelFunc
					var in <-chan Result
					var results = make([]Result, 0)
					var wg sync.WaitGroup

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

			Convey("calling Name should return the expected name", func() {
				So(merged.Name(), ShouldEqual, sourceMergedName)
			})
		})
	})
}

func TestNewMergedSource(t *testing.T) {
	Convey("When calling newMergedSource", t, func() {
		Convey("with a nil Source array", func() {
			var err error
			var merged Source

			merged, err = newMergedSource(nil)

			Convey("it should return an error", func() {
				So(merged, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errSourceNone), ShouldBeTrue)
			})
		})

		Convey("with an empty Source array", func() {
			var err error
			var merged Source

			merged, err = newMergedSource([]Source{})

			Convey("it should return an error", func() {
				So(merged, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errSourceNone), ShouldBeTrue)
			})
		})

		Convey("with a single Source", func() {
			var err error
			var merged Source
			var source Source

			source, err = NewSource(SourceConfig{}, &memFilesystem{})

			So(err, ShouldBeNil)
			So(source, ShouldNotBeNil)

			merged, err = newMergedSource([]Source{source})

			Convey("it should return the same Source", func() {
				So(err, ShouldBeNil)
				So(merged, ShouldNotBeNil)
				So(merged, ShouldEqual, source)
			})
		})

		Convey("with duplicate Sources", func() {
			var err error
			var merged Source
			var source1 Source
			var source2 Source

			source1, err = NewSource(SourceConfig{}, &memFilesystem{
				root: &memFilesystemNode{
					children: map[string]*memFilesystemNode{
						"file1": {},
					},
				},
			})

			So(err, ShouldBeNil)
			So(source1, ShouldNotBeNil)

			source2, err = NewSource(SourceConfig{}, &memFilesystem{
				root: &memFilesystemNode{
					children: map[string]*memFilesystemNode{
						"file2": {},
					},
				},
			})

			So(err, ShouldBeNil)
			So(source2, ShouldNotBeNil)

			merged, err = newMergedSource([]Source{source1, source2, source1, source2})

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
	Convey("When calling NewSource", t, func() {
		Convey("with a nil Filesystem", func() {
			Convey("it should return an error", func() {
				var err error
				var source Source

				source, err = NewSource(SourceConfig{}, nil)

				So(source, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(errors.Is(err, errSourceNilFilesystem), ShouldBeTrue)
			})
		})

		Convey("with a valid Filesystem", func() {
			Convey("it should not return an error", func() {
				var err error
				var source Source

				source, err = NewSource(SourceConfig{}, &memFilesystem{})

				So(err, ShouldBeNil)
				So(source, ShouldNotBeNil)
			})
		})
	})
}

func TestSource(t *testing.T) {
	Convey("When creating a Source", t, func() {
		Convey("which fails immediately upon access and has a Filesystem which performs no action when destroyed",
			func() {
				var err error
				var source Source

				source, err = NewSource(SourceConfig{}, &memFilesystem{
					absolutePathError: errors.New("absolutePath"),
				})

				So(err, ShouldBeNil)
				So(source, ShouldNotBeNil)

				Convey("calling destroy should not perform any action", func() {
					So(source.destroy(), ShouldBeNil)
				})

				Convey("calling Files should return an error", func() {
					var results = collectSourceResults(source)

					So(results, ShouldHaveLength, 1)
					So(results[0].File(), ShouldBeNil)
					So(results[0].Error(), ShouldNotBeNil)
					So(results[0].Error().Error(), ShouldEqual, "absolutePath")
				})
			})

		Convey("which contains a number of files and has a Filesystem which performs an action when destroyed", func() {
			var err error
			var source Source

			source, err = NewSource(SourceConfig{
				Name:    "source",
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
			})

			Convey("calling Files", func() {
				Convey("should return the expected Results", func() {
					var results = collectSourceResults(source)

					So(results, ShouldHaveLength, 3)

					expectFilePathsInResults(nil, results, []string{"file", "dir1/file1", "dir2/file2"})
				})

				Convey("and cancelling should return the expected Results", func(c C) {
					var cancel CancelFunc
					var in <-chan Result
					var results = make([]Result, 0)
					var wg sync.WaitGroup

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
				})
			})

			Convey("calling Name should return the expected name", func() {
				So(source.Name(), ShouldEqual, "source")
			})
		})

		Convey("which panics when calling Files", func() {
			var err error
			var source Source

			source, err = NewSource(SourceConfig{}, &memFilesystem{
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
			})
		})
	})
}

//
// Private functions
//

func collectSourceResults(source Source) []Result {
	var in <-chan Result
	var results = make([]Result, 0)

	in, _ = source.Files(NewContext(ContextConfig{}))

	for result := range in {
		results = append(results, result)
	}

	return results
}

func expectFilePathsInResults(c C, results []Result, paths []string) {
	var pathMap = make(map[string]bool)

	for _, path := range paths {
		pathMap[path] = true
	}

	if c == nil {
		for _, result := range results {
			So(result.Error(), ShouldBeNil)
			So(pathMap, ShouldContainKey, result.File().Path().String())
		}
	} else {
		for _, result := range results {
			c.So(result.Error(), ShouldBeNil)
			c.So(pathMap, ShouldContainKey, result.File().Path().String())
		}
	}
}
