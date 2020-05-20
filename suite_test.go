package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"os"
	"testing"

	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.handcraftedbits.com/pipewerx/internal/event"
)

//
// Public functions
//

func TestMain(m *testing.M) {
	event.AllowFrom(componentFile, true)
	event.AllowFrom(componentFilter, true)
	event.AllowFrom(componentSource, true)

	os.Exit(m.Run())
}

//
// Testcases
//

func TestSuiteCore(t *testing.T) {
	RegisterFailHandler(g.Fail)

	g.RunSpecs(t, "core")
}
