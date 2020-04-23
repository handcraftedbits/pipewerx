package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"fmt"
	"io"
	"os"
	"strings"
)

//
// Public types
//

type File interface {
	Path() FilePath

	Reader() (io.ReadCloser, error)
}

type FileEvaluator interface {
	Destroy() error

	ShouldKeep(file File) (bool, error)
}

type FilePath interface {
	fmt.Stringer

	Dir() []string

	Extension() string

	Name() string
}

// Filesystem is used to abstract filesystem operations and attributes.
type Filesystem interface {
	AbsolutePath(path string) (string, error)

	BasePart(path string) string

	Destroy() error

	DirPart(path string) []string

	ListFiles(path string) ([]os.FileInfo, error)

	PathSeparator() string

	ReadFile(path string) (io.ReadCloser, error)

	StatFile(path string) (os.FileInfo, error)
}

type NewFileEvaluatorFunc func(Context) (FileEvaluator, error)

type Result interface {
	Error() error

	File() File
}

//
// Public functions
//

func NewFilePath(dir []string, name, separator string) FilePath {
	var extension = ""
	var index int

	if dir == nil {
		dir = []string{}
	}

	index = strings.LastIndexByte(name, '.')

	if index != -1 {
		extension = name[index+1:]
		name = name[:index]
	}

	return &filePath{
		dir:       dir,
		extension: extension,
		name:      name,
		separator: separator,
	}
}

//
// Private types
//

// File implementation
type file struct {
	fs   Filesystem
	path FilePath
}

func (f *file) Path() FilePath {
	return f.path
}

func (f *file) Reader() (io.ReadCloser, error) {
	return f.fs.ReadFile(f.path.String())
}

// FilePath implementation
type filePath struct {
	dir       []string
	extension string
	name      string
	separator string
}

func (path *filePath) Dir() []string {
	return path.dir
}

func (path *filePath) Extension() string {
	return path.extension
}

func (path *filePath) Name() string {
	return path.name
}

func (path *filePath) String() string {
	var result string

	if len(path.dir) == 0 {
		result = path.name
	} else {
		result = strings.Join(path.dir, path.separator) + path.separator + path.name
	}

	if path.extension != "" {
		result += "." + path.extension
	}

	return result
}

// FileEvaluator implementation that never excludes files
type nilFileEvaluator struct {
}

func (evaluator *nilFileEvaluator) Destroy() error {
	return nil
}

func (evaluator *nilFileEvaluator) ShouldKeep(file File) (bool, error) {
	return true, nil
}

// Result implementation
type result struct {
	err  error
	file File
}

func (res *result) Error() error {
	return res.err
}

func (res *result) File() File {
	return res.file
}
