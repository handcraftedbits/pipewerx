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
		realPath: producer.stepper.root + producer.stepper.fs.pathSeparator() + path.String(),
	}, nil
}

// filesystem implementation for local filesystem
type localFilesystem struct {
}

func (fs *localFilesystem) absolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

func (fs *localFilesystem) basePart(path string) string {
	return filepath.Base(path)
}

func (fs *localFilesystem) dirPart(path string) []string {
	var dir = filepath.Dir(path)

	if dir == "." {
		// This will be the case for a single file with no directory component, so return an empty array in this case.

		return []string{}
	}

	return strings.Split(filepath.Dir(path), localFSSeparator)
}

func (fs *localFilesystem) listFiles(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}

func (fs *localFilesystem) pathSeparator() string {
	return localFSSeparator
}

func (fs *localFilesystem) statFile(path string) (os.FileInfo, error) {
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
