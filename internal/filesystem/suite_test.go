package filesystem // import "golang.handcraftedbits.com/pipewerx/filesystem"

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.handcraftedbits.com/pipewerx/internal/testutil"
)

//
// Testcases
//

func TestSuiteFilesystem(t *testing.T) {
	RegisterFailHandler(Fail)

	docker = testutil.NewDocker("")
	portSamba = testutil.StartSambaContainer(docker, testutil.TestdataPathFilesystem)

	RunSpecs(t, "filesystem")

	docker.Destroy()
}

//
// Private variables
//

var (
	docker *testutil.Docker

	portSamba int
)
