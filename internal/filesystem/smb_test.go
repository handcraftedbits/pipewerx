package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"io"
	"os"
	"testing"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// SMB filesystem tests

func TestNewSMB(t *testing.T) {
	Convey("When calling SMB", t, func() {
		Convey("it should return an error if an error occurs while creating the Samba context", func() {
			var err error
			var fs pipewerx.Filesystem

			fs, err = SMB(SMBConfig{
				EnableTestConditions: true,
			})

			So(fs, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestSMB(t *testing.T) {
	var docker = testutil.NewDocker("")
	var port int

	defer docker.Destroy()

	Convey("Starting Samba Docker container should succeed", t, func() {
		port = testutil.StartSambaContainer(docker, testutil.TestdataPathFilesystem)
	})

	Convey("When creating an SMB Filesystem with test conditions enabled", t, func() {
		var err error
		var fs pipewerx.Filesystem
		var ok bool
		var smbFS *smb

		fs, err = SMB(newSMBConfig(port))

		So(err, ShouldBeNil)
		So(fs, ShouldNotBeNil)

		smbFS, ok = fs.(*smb)

		So(ok, ShouldBeTrue)

		smbFS.config.EnableTestConditions = true

		defer func() {
			smbFS.config.EnableTestConditions = false

			err = fs.Destroy()

			So(err, ShouldBeNil)
		}()

		Convey("calling Destroy on an SMB Filesystem should return an error", func() {
			err = fs.Destroy()

			So(err, ShouldNotBeNil)
		})

		Convey("calling ListFiles should return an error", func() {
			var fileInfos []os.FileInfo

			fileInfos, err = fs.ListFiles("filesOnly")

			So(fileInfos, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Calling ListFiles on an SMB Filesystem should return an error if an error occurs", t, func() {
		var err error
		var fs pipewerx.Filesystem
		var ok bool
		var smbFS *smb

		fs, err = SMB(SMBConfig{})

		So(err, ShouldBeNil)
		So(fs, ShouldNotBeNil)

		smbFS, ok = fs.(*smb)

		So(ok, ShouldBeTrue)

		smbFS.config.EnableTestConditions = true

		err = fs.Destroy()

		So(err, ShouldNotBeNil)

		smbFS.config.EnableTestConditions = false

		err = fs.Destroy()

		So(err, ShouldBeNil)
	})

	testFilesystem(t, testFilesystemConfig{
		createFunc: func() (pipewerx.Filesystem, error) {
			return SMB(newSMBConfig(port))
		},
		name: "an SMB",
		realPath: func(root, path string) string {
			return path
		},
	})
}

// smbReadCloser tests

func TestSMBReadCloser(t *testing.T) {
	Convey("When creating an smbReadCloser with no file handle", t, func() {
		var err error
		var fs pipewerx.Filesystem
		var ok bool
		var reader io.ReadCloser
		var smbFS *smb

		fs, err = SMB(SMBConfig{})

		So(err, ShouldBeNil)
		So(fs, ShouldNotBeNil)

		smbFS, ok = fs.(*smb)

		So(ok, ShouldBeTrue)

		reader = &smbReadCloser{
			cContext:    smbFS.cContext,
			cFileHandle: nil,
		}

		defer func() {
			err = fs.Destroy()

			So(err, ShouldBeNil)
		}()

		Convey("calling Close should return an error", func() {
			So(reader.Close(), ShouldNotBeNil)
		})

		Convey("calling Read", func() {
			var amountRead int

			Convey("with a nil or empty byte array should return zero bytes read and no error", func() {
				amountRead, err = reader.Read(nil)

				So(amountRead, ShouldEqual, 0)
				So(err, ShouldBeNil)

				amountRead, err = reader.Read([]byte{})

				So(amountRead, ShouldEqual, 0)
				So(err, ShouldBeNil)
			})

			Convey("with a valid byte array should return an error", func() {
				amountRead, err = reader.Read(make([]byte, 10))

				So(amountRead, ShouldBeLessThan, 0)
				So(err, ShouldNotBeNil)
			})
		})
	})
}
