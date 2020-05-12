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
