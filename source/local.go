package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.handcraftedbits.com/pipewerx"
)

//
// Public types
//

type LocalConfig struct {
	Name    string
	Recurse bool
	Root    string
}

//
// Public functions
//

func NewLocal(config *LocalConfig) pipewerx.Source {
	return pipewerx.NewSource(config.Name, func(pipewerx.Context) (pipewerx.FileProducer, error) {
		var err error
		var stepper *pathStepper

		stepper, err = newPathStepper(localFSInstance, config.Root, config.Recurse)

		if err != nil {
			return nil, err
		}

		return &localFileProducer{
			stepper: stepper,
		}, nil
	})
}

//
// Private types
//

// Local filesystem pipewerx.File implementation
type localFile struct {
	path     pipewerx.FilePath
	realPath string
}

func (file *localFile) Path() pipewerx.FilePath {
	return file.path
}

func (file *localFile) Reader() (io.ReadCloser, error) {
	return os.Open(file.realPath)
}

// Local filesystem pipewerx.FileProducer implementation
type localFileProducer struct {
	stepper *pathStepper
}

func (producer *localFileProducer) Destroy() error {
	return nil
}

func (producer *localFileProducer) Next() (pipewerx.File, error) {
	var err error
	var path pipewerx.FilePath

	path, err = producer.stepper.nextFile()

	if err != nil {
		return nil, err
	}

	if path == nil {
		return nil, nil
	}

	return &localFile{
		path:     path,
		realPath: producer.stepper.root + producer.stepper.fs.PathSeparator() + path.String(),
	}, nil
}

// Local filesystem client.Filesystem implementation
type localFilesystem struct {
}

func (fs *localFilesystem) AbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

func (fs *localFilesystem) BasePart(path string) string {
	return filepath.Base(path)
}

func (fs *localFilesystem) DirPart(path string) []string {
	var dir = filepath.Dir(path)

	if dir == "." {
		// This will be the case for a single file with no directory component, so return an empty array.

		return []string{}
	}

	return strings.Split(dir, localFSSeparator)
}

func (fs *localFilesystem) ListFiles(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}

func (fs *localFilesystem) PathSeparator() string {
	return localFSSeparator
}

func (fs *localFilesystem) StatFile(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

//
// Private constants
//

const localFSSeparator = string(os.PathSeparator)

//
// Private variables
//

var localFSInstance = &localFilesystem{}
