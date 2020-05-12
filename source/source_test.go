package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	. "github.com/onsi/ginkgo"
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"
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

func mustCreateSource(config testSourceConfig, id, root string, recurse bool) pipewerx.Source {
	var err error
	var source pipewerx.Source

	source, err = config.createFunc(id, root, recurse)

	Expect(err).To(BeNil())
	Expect(source).NotTo(BeNil())

	return source
}

func testSource(config testSourceConfig) bool {
	return g.Describe(config.name+" Source", func() {
		Describe("calling "+config.name, func() {
			Context("with an invalid ID", func() {
				It("should return an error", func() {
					var err error
					var ids = []string{"", " ", ".", "a ", " a", "a.", ".a", "a..b", "a-b", "?"}
					var source pipewerx.Source

					for _, id := range ids {
						source, err = config.createFunc(id, "", false)

						Expect(source).To(BeNil())
						Expect(err).NotTo(BeNil())
					}
				})
			})
		})

		Describe("given a new instance", func() {
			var results []pipewerx.Result
			var root string
			var sourceName = "source"

			Context("when an empty directory is used as the root", func() {
				BeforeEach(func() {
					root = config.realPath(testutil.TestdataPathFilesystem, "emptyDir")
				})

				Context("and recursion is enabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{}, nil))
						})
					})
				})

				Context("and recursion is disabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{}, nil))
						})
					})
				})
			})

			Context("when a directory containing multiple empty directories is used as the root", func() {
				BeforeEach(func() {
					root = config.realPath(testutil.TestdataPathFilesystem, "multipleEmptyDirs")
				})

				Context("and recursion is enabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{}, nil))
						})
					})
				})

				Context("and recursion is disabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{}, nil))
						})
					})
				})
			})

			Context("when a single file is used as the root", func() {
				BeforeEach(func() {
					root = config.realPath(testutil.TestdataPathFilesystem, "fileOnly.test")
				})

				Context("and recursion is enabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{"fileOnly.test"},
								[]string{"fileOnly"}))
						})
					})
				})

				Context("and recursion is disabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{"fileOnly.test"},
								[]string{"fileOnly"}))
						})
					})
				})
			})

			Context("when there are no subdirectories", func() {
				BeforeEach(func() {
					root = config.realPath(testutil.TestdataPathFilesystem, "filesOnly")
				})

				Context("and recursion is enabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{"a.test", "b.test",
								"c.test"}, []string{"a", "b", "c"}))
						})
					})
				})

				Context("and recursion is disabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{"a.test", "b.test",
								"c.test"}, []string{"a", "b", "c"}))
						})
					})
				})
			})

			Context("when there is a single level of subdirectories", func() {
				BeforeEach(func() {
					root = config.realPath(testutil.TestdataPathFilesystem, "singleLevelSubdirs")
				})

				Context("and recursion is enabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{"a/a.test", "b/b.test",
								"c/c.test"}, []string{"a", "b", "c"}))
						})
					})
				})

				Context("and recursion is disabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{}, nil))
						})
					})
				})
			})

			Context("when there are multiple levels of subdirectories", func() {
				BeforeEach(func() {
					root = config.realPath(testutil.TestdataPathFilesystem, "multiLevelSubdirs")
				})

				Context("and recursion is enabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{"a/a.test", "b/c/c.test",
								"d/e/f/f.test"}, []string{"a", "c", "f"}))
						})
					})
				})

				Context("and recursion is disabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{}, nil))
						})
					})
				})
			})

			Context("when there are multiple levels of subdirectories", func() {
				BeforeEach(func() {
					root = config.realPath(testutil.TestdataPathFilesystem, "mixed")
				})

				Context("and recursion is enabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, true))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{"a.test", "b.test",
								"c/c.test", "d/e/f/f.test"}, []string{"a", "b", "c", "f"}))
						})
					})
				})

				Context("and recursion is disabled", func() {
					Describe("calling Files", func() {
						It("should return the expected Results", func() {
							results = collectSourceResults(mustCreateSource(config, sourceName, root, false))

							Expect(results).To(haveTheseFiles(config.pathSeparator, []string{"a.test", "b.test"},
								[]string{"a", "b"}))
						})
					})
				})
			})
		})
	})
}
