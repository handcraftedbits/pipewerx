package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"testing"
	"time"

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
		port = startSambaContainer()
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

func startSambaContainer() int {
	return testutil.StartSambaContainer(docker, testutil.TestdataPathFilesystem, func(hostPort int) error {
		var err error
		var source pipewerx.Source

		source, err = NewSMB(newSMBConfig(hostPort))

		if err != nil {
			return err
		}

		// libsmbclient doesn't open a connection upon creation, so we'll have to do a simple operation to test for
		// readiness.

		// TODO: not great, find a better way.

		time.Sleep(1 * time.Second)

		if source == nil {
			return nil
		}

		//_, err = source.StatFile("")

		return err
	})
}
