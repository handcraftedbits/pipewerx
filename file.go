package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"fmt"
	"io"
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

type FileProducer interface {
	Destroy() error

	Next() (File, error)
}

type NewFileEvaluatorFunc func(Context) (FileEvaluator, error)

type NewFileProducerFunc func(Context) (FileProducer, error)

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

// FileProducer implementation that returns no results
type nilFileProducer struct {
}

func (producer *nilFileProducer) Destroy() error {
	return nil
}

func (producer *nilFileProducer) Next() (File, error) {
	return nil, nil
}
