package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"testing"

	"golang.handcraftedbits.com/pipewerx"
)

//
// Testcases
//

// Local filesystem tests

func TestLocal(t *testing.T) {
	testFilesystem(t, testFilesystemConfig{
		createFunc: func() (pipewerx.Filesystem, error) {
			return Local(""), nil
		},
		name: "a local",
		realPath: func(root, path string) string {
			return root + localFSSeparator + path
		},
	})
}
