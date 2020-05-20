package event // import "golang.handcraftedbits.com/pipewerx/internal/event"

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Testcases
//

func TestSuiteEvent(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "event")
}
