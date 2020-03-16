package client // import "golang.handcraftedbits.com/pipewerx/internal/client"

import (
	"os"
	"testing"

	"golang.handcraftedbits.com/pipewerx/internal/testutil"
)

//
// Public functions
//

func TestMain(m *testing.M) {
	var code int

	docker = testutil.NewDocker("")

	code = m.Run()

	docker.Destroy()

	os.Exit(code)
}

//
// Private variables
//

var docker *testutil.Docker
