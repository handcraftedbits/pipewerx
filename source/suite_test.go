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

	RunSpecs(t, "source")
}

//
// Suite helpers
//

var _ = BeforeSuite(func() {
	portSamba = testutil.StartSambaContainer(docker, testutil.TestdataPathFilesystem)
})

var _ = AfterSuite(func() {
	docker.Destroy()
})

//
// Private variables
//

var (
	docker = testutil.NewDocker("")

	portSamba int
)
