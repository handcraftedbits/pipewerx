package testutil // import "golang.handcraftedbits.com/pipewerx/internal/testutil"

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Testcases
//

func TestSuiteTestutil(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "testutil")
}

//
// Suite helpers
//

var _ = BeforeSuite(func() {
	docker = NewDocker("")
})

var _ = AfterSuite(func() {
	docker.Destroy()
})

//
// Private variables
//

var docker *Docker
