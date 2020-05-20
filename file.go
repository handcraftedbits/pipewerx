package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"fmt"
	"io"
	"os"
	pathutil "path"
	"strings"
	"time"

	"golang.handcraftedbits.com/pipewerx/internal/event"
)

//
// Public types
//

type File interface {
	os.FileInfo

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

// Filesystem is used to abstract filesystem operations and properties.
type Filesystem interface {
	AbsolutePath(path string) (string, error)

	BasePart(path string) string

	Destroy() error

	DirPart(path string) []string

	// ListFiles lists information for all the files directly under a given path (i.e., if the path represents a
	// directory, it will return information for all the files in the directory without entering subdirectories; if the
	// path represents a file, only information for that file will be returned).  Note that the path is considered
	// absolute in the sense that it starts with the root path for the Source that uses this Filesystem.
	ListFiles(path string) ([]os.FileInfo, error)

	PathSeparator() string

	// ReadFile retrieves an io.ReadCloser that can be used to read the contents of a file at a given path.  Note that
	// the path is considered relative in the sense that it does not start with the root path for the Source that uses
	// this Filesystem.
	ReadFile(path string) (io.ReadCloser, error)

	// StatFile retrieves information for a single file at a given path.  Note that the path is considered absolute in
	// the sense that it starts with the root path for the Source that uses this Filesystem.
	StatFile(path string) (os.FileInfo, error)
}

// FilesystemDefaults is used to help implement Filesystems by providing implementations for AbsolutePath, BasePart,
// DirPart, and PathSeparator that are appropriate for most filesystems (namely, those with UNIX-style paths).
type FilesystemDefaults struct {
}

func (fs FilesystemDefaults) AbsolutePath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	return pathutil.Clean(path), nil
}

func (fs FilesystemDefaults) BasePart(path string) string {
	return pathutil.Base(path)
}

func (fs FilesystemDefaults) DirPart(path string) []string {
	var dir = pathutil.Dir(path)

	if dir == "." {
		// This will be the case for a single file with no directory component, so return an empty array.

		return []string{}
	}

	return strings.Split(dir, "/")
}

func (fs FilesystemDefaults) PathSeparator() string {
	return "/"
}

type Result interface {
	Error() error

	File() File
}

//
// Private types
//

// io.ReadCloser implementation that produces Events detailing file read progress.
type eventProducingReadCloser struct {
	file     File
	sourceID string
	wrapped  io.ReadCloser
}

func (reader *eventProducingReadCloser) Read(p []byte) (int, error) {
	var amount int
	var err error

	amount, err = reader.wrapped.Read(p)

	if amount > 0 && event.IsAllowedFrom(componentFile) {
		event.Send(fileEventRead(reader.file, reader.sourceID, amount))
	}

	return amount, err
}

func (reader *eventProducingReadCloser) Close() error {
	if event.IsAllowedFrom(componentFile) {
		event.Send(fileEventClosed(reader.file, reader.sourceID))
	}

	return reader.wrapped.Close()
}

// File implementation
type file struct {
	fileInfo os.FileInfo
	fs       Filesystem
	path     FilePath
	sourceID string
}

func (f *file) IsDir() bool {
	return f.fileInfo.IsDir()
}

func (f *file) Mode() os.FileMode {
	return f.fileInfo.Mode()
}

func (f *file) ModTime() time.Time {
	return f.fileInfo.ModTime()
}

func (f *file) Name() string {
	var name = f.path.Name()

	if f.path.Extension() != "" {
		name += "." + f.path.Extension()
	}

	return name
}

func (f *file) Path() FilePath {
	return f.path
}

func (f *file) Reader() (io.ReadCloser, error) {
	var err error
	var reader io.ReadCloser

	reader, err = f.fs.ReadFile(f.path.String())

	if err != nil {
		return nil, err
	}

	if event.IsAllowedFrom(componentFile) {
		event.Send(fileEventOpened(f, f.sourceID))
	}

	return &eventProducingReadCloser{
		file:     f,
		sourceID: f.sourceID,
		wrapped:  reader,
	}, nil
}

func (f *file) Size() int64 {
	return f.fileInfo.Size()
}

func (f *file) Sys() interface{} {
	return f.fileInfo.Sys()
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

//
// Private functions
//

func fileEventClosed(file File, sourceOrDestID string) event.Event {
	return newFileEvent(event.TypeClosed, file, sourceOrDestID)
}

func fileEventOpened(file File, sourceOrDestID string) event.Event {
	var evt = newFileEvent(event.TypeOpened, file, sourceOrDestID)

	evt.Data()[event.FieldLength] = file.Size()

	return evt
}

func fileEventRead(file File, sourceOrDestID string, amount int) event.Event {
	var evt = newFileEvent(event.TypeRead, file, sourceOrDestID)

	evt.Data()[event.FieldLength] = amount

	return evt
}

func newFileEvent(eventType string, file File, sourceOrDestID string) event.Event {
	var evt = event.WithID(componentFile, sourceOrDestID, eventType)

	evt.Data()[event.FieldFile] = file.Path().String()

	return evt
}

func newFilePath(dir []string, name, separator string) FilePath {
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
