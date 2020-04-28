package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"testing"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// SMB Source tests

func TestNewSMB(t *testing.T) {
	Convey("When calling NewSMB", t, func() {
		Convey("it should return an error when test conditions are enabled", func() {
			var err error
			var source pipewerx.Source

			source, err = NewSMB(SMBConfig{
				enableTestConditions: true,
			})

			So(source, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestSMB(t *testing.T) {
	var port int

	Convey("Starting Samba Docker container should succeed", t, func() {
		port = testutil.StartSambaContainer(docker, testutil.TestdataPathFilesystem)
	})

	testSource(t, testSourceConfig{
		createFunc: func(root string, recurse bool) (pipewerx.Source, error) {
			var config = newSMBConfig(port)

			config.Recurse = recurse
			config.Root = root

			return NewSMB(config)
		},
		name:          "an SMB",
		pathSeparator: "/",
		realPath: func(root, path string) string {
			return path
		},
	})
}

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
