package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
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
	Convey("When calling NewSMB", t, func() {
		Convey("it should return an error when", func() {
			var config SMBConfig
			var err error
			var fs pipewerx.Filesystem

			config = newSMBConfig(startSambaContainer())

			Convey("an invalid host is provided", func() {
				config.Host = "????"

				fs, err = NewSMB(config)

				So(fs, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("an invalid password is provided", func() {
				config.Password = "????"

				fs, err = NewSMB(config)

				So(fs, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("an invalid port is provided", func() {
				config.Port = -1

				fs, err = NewSMB(config)

				So(fs, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("an invalid share is provided", func() {
				config.Share = "????"

				fs, err = NewSMB(config)

				So(fs, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("an invalid username is provided", func() {
				config.Username = "????"

				fs, err = NewSMB(config)

				So(fs, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestSMB(t *testing.T) {
	var port int

	Convey("Starting Samba Docker container should succeed", t, func() {
		port = startSambaContainer()
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
		_, clientError := NewSMB(newSMBConfig(hostPort))

		return clientError
	})
}
