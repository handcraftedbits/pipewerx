package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"golang.handcraftedbits.com/pipewerx"
)

//
// Testcases
//

// Local filesystem tests

var _ = testFilesystem(testFilesystemConfig{
	createFunc: func() (pipewerx.Filesystem, error) {
		return Local(""), nil
	},
	name: "Local",
	realPath: func(root, path string) string {
		return root + localFSSeparator + path
	},
})
