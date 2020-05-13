package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Testcases
//

// fileInfo tests

var _ = Describe("fileInfo", func() {
	Describe("given a new instance", func() {
		var fi *fileInfo
		var now time.Time

		BeforeEach(func() {
			now = time.Now()
			fi = &fileInfo{
				mode:    os.ModeDir,
				modTime: now,
				name:    "name",
				size:    1,
			}
		})

		Describe("calling IsDir", func() {
			It("should return the expected value", func() {
				Expect(fi.IsDir()).To(BeTrue())
			})
		})

		Describe("calling Mode", func() {
			It("should return the expected mod", func() {
				Expect(fi.Mode()).To(Equal(os.ModeDir))
			})
		})

		Describe("calling ModTime", func() {
			It("should return the expected modification time", func() {
				Expect(fi.ModTime()).To(Equal(now))
			})
		})

		Describe("calling Name", func() {
			It("should return the expected name", func() {
				Expect(fi.Name()).To(Equal("name"))
			})
		})

		Describe("calling Size", func() {
			It("should return the expected size", func() {
				Expect(fi.Size()).To(BeEquivalentTo(1))
			})
		})

		Describe("calling Sys", func() {
			It("should return the expected value", func() {
				Expect(fi.Sys()).To(BeNil())
			})
		})
	})
})
