package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"
)

//
// Testcases
//

// SMB filesystem tests

var _ = Describe("SMB Filesystem", func() {
	Describe("calling SMB", func() {
		It("should return an error if an error occurs while creating the Samba context", func() {
			var err error
			var fs pipewerx.Filesystem

			fs, err = SMB(SMBConfig{
				EnableTestConditions: true,
			})

			Expect(fs).To(BeNil())
			Expect(err).NotTo(BeNil())
		})
	})

	Describe("given a new instance", func() {
		Context("with test conditions enabled", func() {
			var err error
			var fs pipewerx.Filesystem
			var smbFS *smb

			BeforeEach(func() {
				var ok bool

				fs, err = SMB(newSMBConfig(portSamba))

				Expect(err).To(BeNil())
				Expect(fs).NotTo(BeNil())

				smbFS, ok = fs.(*smb)

				Expect(ok).To(BeTrue())

				smbFS.config.EnableTestConditions = true
			})

			AfterEach(func() {
				smbFS.config.EnableTestConditions = false

				err = fs.Destroy()

				Expect(err).To(BeNil())
			})

			Describe("calling Destroy", func() {
				It("should return an error", func() {
					err = fs.Destroy()

					Expect(err).NotTo(BeNil())
				})
			})

			Describe("calling ListFiles", func() {
				It("should return an error", func() {
					var fileInfos []os.FileInfo

					fileInfos, err = fs.ListFiles("filesOnly")

					Expect(fileInfos).To(BeNil())
					Expect(err).NotTo(BeNil())
				})
			})
		})
	})
})

var _ = testFilesystem(testFilesystemConfig{
	createFunc: func() (pipewerx.Filesystem, error) {
		return SMB(newSMBConfig(portSamba))
	},
	name: "SMB",
	realPath: func(root, path string) string {
		return path
	},
})

// smbReadCloser tests

var _ = Describe("smbReadCloser", func() {
	Describe("given a new instance", func() {
		Context("with a nil file handle", func() {
			var err error
			var fs pipewerx.Filesystem
			var reader io.ReadCloser

			BeforeEach(func() {
				var ok bool
				var smbFS *smb

				fs, err = SMB(SMBConfig{})

				Expect(err).To(BeNil())
				Expect(fs).NotTo(BeNil())

				smbFS, ok = fs.(*smb)

				Expect(ok).To(BeTrue())

				reader = &smbReadCloser{
					cContext:    smbFS.cContext,
					cFileHandle: nil,
				}
			})

			AfterEach(func() {
				err = fs.Destroy()

				Expect(err).To(BeNil())
			})

			Describe("calling Close", func() {
				It("should return an error", func() {
					Expect(reader.Close()).NotTo(BeNil())
				})
			})

			Describe("calling Read", func() {
				var amountRead int

				Context("with a nil or empty byte array", func() {
					It("should return zero bytes read and no error", func() {
						amountRead, err = reader.Read(nil)

						Expect(amountRead).To(Equal(0))
						Expect(err).To(BeNil())

						amountRead, err = reader.Read([]byte{})

						Expect(amountRead).To(Equal(0))
						Expect(err).To(BeNil())
					})
				})

				Context("with a valid byte array", func() {
					It("should return an error", func() {
						amountRead, err = reader.Read(make([]byte, 10))

						Expect(amountRead < 0).To(BeTrue())
						Expect(err).NotTo(BeNil())
					})
				})
			})
		})
	})
})

//
// Private functions
//

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
