package source

import (
	"testing"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/client"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

func TestSMBFileProducer_Next(t *testing.T) {
	Convey("When creating a SMB Source", t, func() {
		var port = startSambaContainer()

		testFileProducer("", func(root string, recurse bool) pipewerx.Source {
			var config = newSMBConfig(port)

			config.Name = "smb"
			config.Recurse = recurse
			config.Root = root

			return NewSMB(config)
		}, func(stepper *pathStepper) pipewerx.FileProducer {
			return &smbFileProducer{
				smbClient: newSMBClient(port),
				stepper:   stepper,
			}
		}, func() client.Filesystem {
			return newSMBClient(port)
		})
	})
}

//
// Private functions
//

func newSMBClient(port int) client.SMB {
	smbClient, err := client.NewSMB(newSMBClientConfig(port))

	So(err, ShouldBeNil)

	return smbClient
}

func newSMBClientConfig(port int) *client.SMBConfig {
	return &client.SMBConfig{
		Domain:   testutil.ConstSMBDomain,
		Host:     "localhost",
		Password: testutil.ConstSMBPassword,
		Port:     port,
		Share:    testutil.ConstSMBShare,
		Username: testutil.ConstSMBUser,
	}
}

func newSMBConfig(port int) *SMBConfig {
	var config = newSMBClientConfig(port)

	return &SMBConfig{
		Domain:   config.Domain,
		Host:     config.Host,
		Password: config.Password,
		Port:     config.Port,
		Share:    config.Share,
		Username: config.Username,
	}
}

func startSambaContainer() int {
	return testutil.StartSambaContainer(docker, testDataRoot, func(hostPort int) error {
		smbClient, clientError := client.NewSMB(newSMBClientConfig(hostPort))

		if smbClient != nil {
			smbClient.Disconnect()
		}

		return clientError
	})
}
