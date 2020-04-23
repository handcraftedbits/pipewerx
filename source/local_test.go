package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"os"
	"testing"

	"golang.handcraftedbits.com/pipewerx"
)

//
// Testcases
//

// Local Source tests

func TestLocal(t *testing.T) {
	testSource(t, testSourceConfig{
		createFunc: func(root string, recurse bool) (pipewerx.Source, error) {
			return NewLocal(LocalConfig{
				Recurse: recurse,
				Root:    root,
			})
		},
		name:          "a local",
		pathSeparator: "/",
		realPath: func(root, path string) string {
			return root + string(os.PathSeparator) + path
		},
	})
}
