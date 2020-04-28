package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"io"
	"os"
	"testing"
	"time"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// SMB filesystem tests

func TestNewSMB(t *testing.T) {
	Convey("When calling NewSMB", t, func() {
		Convey("it should return an error if an error occurs while creating the Samba context", func() {
			var err error
			var fs pipewerx.Filesystem

			fs, err = NewSMB(SMBConfig{
				EnableTestConditions: true,
			})

			So(fs, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestSMB(t *testing.T) {
	var port int

	Convey("Starting Samba Docker container should succeed", t, func() {
		port = startSambaContainer()
	})

	Convey("When creating an SMB Filesystem with test conditions enabled", t, func() {
		var err error
		var fs pipewerx.Filesystem
		var ok bool
		var smbFS *smb

		fs, err = NewSMB(newSMBConfig(port))

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

		fs, err = NewSMB(SMBConfig{})

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
			return NewSMB(newSMBConfig(port))
		},
		name: "an SMB",
		realPath: func(root, path string) string {
			return path
		},
	})
}

// smbFileInfo tests

func TestSMBFileInfo(t *testing.T) {
	var now = time.Now()

	Convey("When creating an smbFileInfo", t, func() {
		var fileInfo = &smbFileInfo{
			mode:    os.ModeDir,
			modTime: now,
			name:    "name",
			size:    1,
		}

		Convey("calling IsDir should return the expected value", func() {
			So(fileInfo.IsDir(), ShouldBeTrue)
		})

		Convey("calling Mode should return the expected value", func() {
			So(fileInfo.Mode(), ShouldEqual, os.ModeDir)
		})

		Convey("calling ModTime should return the expected value", func() {
			So(fileInfo.ModTime(), ShouldEqual, now)
		})

		Convey("calling Name should return the expected value", func() {
			So(fileInfo.Name(), ShouldEqual, "name")
		})

		Convey("calling Size should return the expected value", func() {
			So(fileInfo.Size(), ShouldEqual, 1)
		})

		Convey("calling Sys should return the expected value", func() {
			So(fileInfo.Sys(), ShouldEqual, nil)
		})
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

		fs, err = NewSMB(SMBConfig{})

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

			Convey("with a nil or empty byte array should zero bytes read and no error", func() {
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

//
// Private functions
//

// TODO: maybe move into testutil... ?

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

func startSambaContainer() int {
	return testutil.StartSambaContainer(docker, testutil.TestdataPathFilesystem, func(hostPort int) error {
		var err error
		var fs pipewerx.Filesystem

		fs, err = NewSMB(newSMBConfig(hostPort))

		if err != nil {
			return err
		}

		// libsmbclient doesn't open a connection upon creation, so we'll have to do a simple operation to test for
		// readiness.

		_, err = fs.StatFile("")

		return err
	})
}
