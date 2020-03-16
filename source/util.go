package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"os"
	"strings"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/client"
)

//
// Private types
//

// pathStepper is used to "step" through a filesystem path by listing one file at a time.
type pathStepper struct {
	dirs  *stringStack
	files *stringStack
	fs    client.Filesystem
	root  string
}

func (stepper *pathStepper) nextFile() (pipewerx.FilePath, error) {
	for stepper.files.isEmpty() {
		if stepper.dirs.isEmpty() {
			return nil, nil
		}

		if err := findFiles(stepper.fs, stepper.dirs.pop(), stepper.dirs, stepper.files); err != nil {
			return nil, err
		}
	}

	return newFilePathFromString(stepper.fs, stepper.root, stepper.files.pop(), stepper.fs.PathSeparator()), nil
}

// stringStack is used to create a simple stack of strings.
type stringStack []string

func (stack *stringStack) clear() {
	*stack = nil
}

func (stack *stringStack) isEmpty() bool {
	return len(*stack) == 0
}

func (stack *stringStack) peek() string {
	return (*stack)[len(*stack)-1]
}

func (stack *stringStack) pop() string {
	var end = len(*stack) - 1
	var item = (*stack)[end]

	*stack = (*stack)[:end]

	return item
}

func (stack *stringStack) push(item string) {
	*stack = append(*stack, item)
}

//
// Private functions
//

// Finds all files and directories within the current path, without diving into subdirectories.
func findFiles(fs client.Filesystem, path string, dirs, files *stringStack) error {
	var err error
	var fileInfo os.FileInfo
	var fileInfos []os.FileInfo

	fileInfo, err = fs.StatFile(path)

	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		files.push(path)

		return nil
	}

	fileInfos, err = fs.ListFiles(path)

	if err != nil {
		return err
	}

	for _, fileInfo = range fileInfos {
		var newPath = path + fs.PathSeparator() + fileInfo.Name()

		if fileInfo.IsDir() {
			dirs.push(newPath)
		} else {
			files.push(newPath)
		}
	}

	return nil
}

func newFilePathFromString(fs client.Filesystem, root, path, separator string) pipewerx.FilePath {
	path = stripRoot(root, path, separator)

	return pipewerx.NewFilePath(fs.DirPart(path), fs.BasePart(path), separator)
}

func newPathStepper(fs client.Filesystem, root string, recurse bool) (*pathStepper, error) {
	var err error
	var dirs = &stringStack{}
	var files = &stringStack{}

	root, err = fs.AbsolutePath(root)

	if err != nil {
		return nil, err
	}

	if err = findFiles(fs, root, dirs, files); err != nil {
		return nil, err
	}

	if !recurse {
		// Prevents us from going into any subdirectories (i.e., we won't have any to visit).

		dirs.clear()
	}

	if !files.isEmpty() && (root == files.peek()) {
		// Special case.  This implies that the root is a single file, not a directory, so we need to adjust the root
		// directory accordingly or else downstream methods like File.Reader() will fail.

		root = strings.Join(fs.DirPart(root), fs.PathSeparator())
	}

	return &pathStepper{
		dirs:  dirs,
		files: files,
		fs:    fs,
		root:  root,
	}, nil
}

func stripRoot(root, path, separator string) string {
	path = path[len(root):]

	if strings.HasPrefix(path, separator) {
		return path[len(separator):]
	}

	return path
}
