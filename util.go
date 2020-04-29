package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

//
// Public types
//

type CancelFunc func(func())

//
// Private types
//

// TODO: mention others as they are written.
// cancellationHelper is used to simplify the handling of cancellation events in Sources and Filters.
type cancellationHelper struct {
	callback  func()
	cancel    chan<- struct{}
	cancelled bool
	logger    *zerolog.Logger
	mutex     sync.Mutex
	out       chan<- Result
	wg        *sync.WaitGroup
}

func (helper *cancellationHelper) finalize() {
	if helper.wg != nil {
		helper.wg.Wait()
	}

	func() {
		defer func() {
			if value := recover(); value != nil {
				helper.logPanic(value)
			}
		}()

		close(helper.out)
	}()

	helper.mutex.Lock()
	defer helper.mutex.Unlock()

	if !helper.cancelled {
		helper.cancelled = true

		func() {
			defer func() {
				if value := recover(); value != nil {
					helper.logPanic(value)
				}
			}()

			close(helper.cancel)
		}()
	}

	if helper.callback != nil {
		func() {
			defer func() {
				if value := recover(); value != nil {
					helper.logPanic(value)
				}
			}()

			helper.callback()
		}()
	}
}

func (helper *cancellationHelper) invoker() CancelFunc {
	return func(callback func()) {
		go func() {
			helper.mutex.Lock()
			defer helper.mutex.Unlock()

			if !helper.cancelled {
				helper.cancelled = true
				helper.callback = callback

				close(helper.cancel)
			}
		}()
	}
}

func (helper *cancellationHelper) logPanic(value interface{}) {
	helper.logger.Warn().
		Interface("error", value).
		Msg("an unexpected error occurred during cancellation")
}

// pathStepper is used to "step" through a filesystem path by listing one file at a time.
type pathStepper struct {
	dirs  *stringStack
	files *stepperFileStack
	fs    Filesystem
	root  string
}

func (stepper *pathStepper) nextFile() (File, error) {
	var curFile *stepperFile

	for stepper.files.isEmpty() {
		if stepper.dirs.isEmpty() {
			return nil, nil
		}

		if err := findFiles(stepper.fs, stepper.dirs.pop(), stepper.dirs, stepper.files); err != nil {
			return nil, err
		}
	}

	curFile = stepper.files.pop()

	return &file{
		fileInfo: curFile.fileInfo,
		fs:       stepper.fs,
		path:     newFilePathFromString(stepper.fs, stepper.root, curFile.path),
	}, nil
}

// stepperFile is used to capture information about a file encountered by a pathStepper.
type stepperFile struct {
	fileInfo os.FileInfo
	path     string
}

// stepperFileStack is used to create a stack of stepperFiles.
type stepperFileStack []*stepperFile

func (stack *stepperFileStack) clear() {
	*stack = nil
}

func (stack *stepperFileStack) isEmpty() bool {
	return len(*stack) == 0
}

func (stack *stepperFileStack) peek() *stepperFile {
	return (*stack)[len(*stack)-1]
}

func (stack *stepperFileStack) pop() *stepperFile {
	var end = len(*stack) - 1
	var item = (*stack)[end]

	*stack = (*stack)[:end]

	return item
}

func (stack *stepperFileStack) push(item *stepperFile) {
	*stack = append(*stack, item)
}

// stringStack is used to create a stack of strings.
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
func findFiles(fs Filesystem, path string, dirs *stringStack, files *stepperFileStack) (e error) {
	var err error
	var fileInfo os.FileInfo
	var fileInfos []os.FileInfo

	defer func() {
		if value := recover(); value != nil {
			e = newPanicError(value)
		}
	}()

	fileInfo, err = fs.StatFile(path)

	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		files.push(&stepperFile{
			fileInfo: fileInfo,
			path:     path,
		})

		return nil
	}

	fileInfos, err = fs.ListFiles(path)

	if err != nil {
		return err
	}

	for _, fileInfo = range fileInfos {
		var newPath = path

		// If the current path is the top of the filesystem (e.g., "/"), we don't want to add an extra path separator,
		// since that would be redundant.

		if newPath != fs.PathSeparator() {
			newPath += fs.PathSeparator()
		}

		newPath += fileInfo.Name()

		if fileInfo.IsDir() {
			dirs.push(newPath)
		} else {
			files.push(&stepperFile{
				fileInfo: fileInfo,
				path:     newPath,
			})
		}
	}

	return nil
}

func newCancellationHelper(logger *zerolog.Logger, out chan<- Result, cancel chan<- struct{},
	wg *sync.WaitGroup) *cancellationHelper {
	return &cancellationHelper{
		cancel: cancel,
		logger: logger,
		mutex:  sync.Mutex{},
		out:    out,
		wg:     wg,
	}
}

func newFilePathFromString(fs Filesystem, root, path string) FilePath {
	// TODO: see if there's a way around this.  Seems unnecessary.
	//  When StatFile or ListFiles is called, the path includes the root.  Otherwise it doesn't.  So why do we need
	//  to include and strip the root when returning a file, which is only going to be passed to ReadFile()?
	//  It doesn't expect the path to include the root.
	//  Also, add comments in Filesystem that explains what kind of path is passed to StatFile, ListFiles, and ReadFile.
	path = stripRoot(root, path, fs.PathSeparator())

	return NewFilePath(fs.DirPart(path), fs.BasePart(path), fs.PathSeparator())
}

func newPathStepper(fs Filesystem, root string, recurse bool) (p *pathStepper, e error) {
	var err error
	var dirs = &stringStack{}
	var files = &stepperFileStack{}

	defer func() {
		if value := recover(); value != nil {
			e = newPanicError(value)
			p = nil
		}
	}()

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

	if !files.isEmpty() && (root == files.peek().path) {
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
