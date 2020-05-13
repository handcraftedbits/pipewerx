package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"os"
	"testing"

	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Public functions
//

func TestMain(m *testing.M) {
	allowEventsFrom(componentFile, true)
	allowEventsFrom(componentFilter, true)
	allowEventsFrom(componentSource, true)

	os.Exit(m.Run())
}

//
// Testcases
//

func TestSuiteCore(t *testing.T) {
	RegisterFailHandler(g.Fail)

	g.RunSpecs(t, "core")
}
