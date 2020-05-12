package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"
)

//
// Private types
//

type testFilesystemConfig struct {
	createFunc func() (pipewerx.Filesystem, error)
	name       string
	realPath   func(string, string) string
}

//
// Private functions
//

func testFilesystem(config testFilesystemConfig) bool {
	return Describe(config.name+" Filesystem", func() {
		Describe("given a new instance", func() {
			var err error
			var fs pipewerx.Filesystem
			var result string
			var sep string

			BeforeEach(func() {
				fs, err = config.createFunc()

				Expect(err).To(BeNil())
				Expect(fs).NotTo(BeNil())

				sep = fs.PathSeparator()
			})

			AfterEach(func() {
				err = fs.Destroy()

				Expect(err).To(BeNil())
			})

			Describe("calling AbsolutePath", func() {
				It("should return the correct absolute path", func() {
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

						Expect(err).To(BeNil())
						Expect(result).To(HaveSuffix(test.expected))
					}
				})
			})

			Describe("calling BasePart", func() {
				It("should return the correct base part", func() {
					var tests = []struct {
						input    string
						expected string
					}{
						{"abc", "abc"},
						{fmt.Sprintf("a%sb%sc%sabc", sep, sep, sep), "abc"},
					}

					for _, test := range tests {
						result = fs.BasePart(test.input)

						Expect(result).To(Equal(test.expected))
					}
				})
			})

			Describe("calling DirPart", func() {
				It("should return the correct directory part", func() {
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

						Expect(result).To(HaveLen(len(test.expected)))

						for i, value := range result {
							Expect(result[i]).To(Equal(value))
						}
					}
				})
			})

			Describe("calling ListFiles", func() {
				var fileInfos []os.FileInfo

				Context("when an invalid directory is specified", func() {
					It("should return an error", func() {
						fileInfos, err = fs.ListFiles(config.realPath(testutil.TestdataPathFilesystem, ";;;;"))

						Expect(fileInfos).To(BeNil())
						Expect(err).NotTo(BeNil())
					})
				})

				Context("when an empty directory is specified", func() {
					It("should return an empty array", func() {
						fileInfos, err = fs.ListFiles(config.realPath(testutil.TestdataPathFilesystem, "emptyDir"))

						Expect(err).To(BeNil())
						Expect(fileInfos).NotTo(BeNil())
						Expect(fileInfos).To(HaveLen(0))
					})
				})

				Context("when a non-empty directory is specified", func() {
					It("should return the expected array of os.FileInfo objects", func() {
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

						Expect(err).To(BeNil())
						Expect(fileInfos).NotTo(BeNil())
						Expect(fileInfos).To(HaveLen(4))

						for _, fileInfo := range fileInfos {
							fileInfosByName[fileInfo.Name()] = fileInfo
						}

						for _, test := range tests {
							var fileInfo = fileInfosByName[test.name]

							Expect(fileInfo).NotTo(BeNil())
							Expect(fileInfo.IsDir()).To(Equal(test.dir))
						}
					})
				})
			})

			Describe("calling ReadFile", func() {
				var reader io.ReadCloser

				Context("when an invalid file is specified", func() {
					It("should return an error", func() {
						reader, err = fs.ReadFile(config.realPath(testutil.TestdataPathFilesystem, ";;;;"))

						Expect(reader).To(BeNil())
						Expect(err).NotTo(BeNil())
					})
				})

				Context("when a valid file is specified", func() {
					It("should succeed", func() {
						var contents []byte

						reader, err = fs.ReadFile(config.realPath(testutil.TestdataPathFilesystem, "fileOnly.test"))

						Expect(err).To(BeNil())
						Expect(reader).NotTo(BeNil())

						contents, err = ioutil.ReadAll(reader)

						Expect(err).To(BeNil())
						Expect(contents).NotTo(BeNil())
						Expect(string(contents)).To(Equal("fileOnly"))

						err = reader.Close()

						Expect(err).To(BeNil())
					})
				})
			})

			Describe("calling StatFile", func() {
				var fileInfo os.FileInfo

				Context("when an invalid file is specified", func() {
					It("should return an error", func() {
						fileInfo, err = fs.StatFile(config.realPath(testutil.TestdataPathFilesystem, ";;;;"))

						Expect(fileInfo).To(BeNil())
						Expect(err).NotTo(BeNil())
					})
				})

				Context("when a valid file is specified", func() {
					It("should succeed", func() {
						fileInfo, err = fs.StatFile(config.realPath(testutil.TestdataPathFilesystem, "fileOnly.test"))

						Expect(err).To(BeNil())
						Expect(fileInfo).NotTo(BeNil())
						Expect(fileInfo.Name()).To(Equal("fileOnly.test"))
						Expect(fileInfo.IsDir()).To(BeFalse())
					})
				})
			})
		})
	})
}
