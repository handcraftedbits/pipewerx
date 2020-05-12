package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"os"

	"golang.handcraftedbits.com/pipewerx"
)

//
// Testcases
//

// Local Source tests

var _ = testSource(testSourceConfig{
	createFunc: func(id, root string, recurse bool) (pipewerx.Source, error) {
		return Local(LocalConfig{
			ID:      id,
			Recurse: recurse,
			Root:    root,
		})
	},
	name:          "Local",
	pathSeparator: "/",
	realPath: func(root, path string) string {
		return root + string(os.PathSeparator) + path
	},
})
