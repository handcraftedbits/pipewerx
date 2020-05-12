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

	RunSpecs(t, "filesystem")
}

//
// Suite helpers
//

var _ = BeforeSuite(func() {
	portSamba = testutil.StartSambaContainer2(docker, testutil.TestdataPathFilesystem)
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
