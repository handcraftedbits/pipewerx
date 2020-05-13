package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.handcraftedbits.com/pipewerx/internal/testutil"
)

//
// Testcases
//

func TestSuiteSource(t *testing.T) {
	RegisterFailHandler(Fail)

	docker = testutil.NewDocker("")
	portSamba = testutil.StartSambaContainer(docker, testutil.TestdataPathFilesystem)

	RunSpecs(t, "source")

	docker.Destroy()
}

//
// Private variables
//

var (
	docker *testutil.Docker

	portSamba int
)
