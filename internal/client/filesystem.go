package client // import "golang.handcraftedbits.com/pipewerx/internal/client"

import (
	"os"
)

//
// Public types
//

// Filesystem is used to abstract filesystem operations and attributes.
type Filesystem interface {
	AbsolutePath(path string) (string, error)

	BasePart(path string) string

	DirPart(path string) []string

	ListFiles(path string) ([]os.FileInfo, error)

	PathSeparator() string

	StatFile(path string) (os.FileInfo, error)
}
