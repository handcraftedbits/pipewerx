package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.handcraftedbits.com/pipewerx"
)

//
// Public functions
//

func Local(root string) pipewerx.Filesystem {
	return &local{
		root: root,
	}
}

//
// Private types
//

// pipewerx.Filesystem implementation for the local filesystem
type local struct {
	root string
}

func (fs *local) AbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

func (fs *local) BasePart(path string) string {
	return filepath.Base(path)
}

func (fs *local) Destroy() error {
	return nil
}

func (fs *local) DirPart(path string) []string {
	var dir = filepath.Dir(path)

	if dir == "." {
		// This will be the case for a single file with no directory component, so return an empty array.

		return []string{}
	}

	return strings.Split(dir, localFSSeparator)
}

func (fs *local) ListFiles(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}

func (fs *local) PathSeparator() string {
	return localFSSeparator
}

func (fs *local) ReadFile(path string) (io.ReadCloser, error) {
	// If the base part of the filesystem root path is the same as the path, that implies that the root is a single file
	// and we can't prepend the filesystem root to the path of the file that we're opening.

	if fs.BasePart(fs.root) != path {
		path = fs.root + localFSSeparator + path
	} else {
		path = fs.root
	}

	return os.Open(path)
}

func (fs *local) StatFile(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

//
// Private constants
//

const localFSSeparator = string(os.PathSeparator)
