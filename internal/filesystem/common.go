package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"os"
	"time"
)

//
// Private types
//

// Simple os.FileInfo implementation
type fileInfo struct {
	mode    os.FileMode
	modTime time.Time
	name    string
	size    int64
}

func (fi *fileInfo) IsDir() bool {
	return fi.mode.IsDir()
}

func (fi *fileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi *fileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *fileInfo) Name() string {
	return fi.name
}

func (fi *fileInfo) Size() int64 {
	return fi.size
}

func (fi *fileInfo) Sys() interface{} {
	return nil
}
