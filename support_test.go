package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Private types
//

// FileEvaluator implementation that discards files based on file extension.
type extensionFileEvaluator struct {
	destroyError    error
	extension       string
	panic           bool
	shouldKeepError error
}

func (evaluator *extensionFileEvaluator) Destroy() error {
	if evaluator.panic {
		panic(evaluator.destroyError)
	}

	return evaluator.destroyError
}

func (evaluator *extensionFileEvaluator) ShouldKeep(file File) (bool, error) {
	if evaluator.shouldKeepError != nil {
		if evaluator.panic {
			panic(evaluator.shouldKeepError)
		}

		return false, evaluator.shouldKeepError
	}

	if file.Path().Extension() == evaluator.extension {
		return true, nil
	}

	return false, nil
}

// In-memory Filesystem implementation.
type memFilesystem struct {
	FilesystemDefaults

	absolutePathError error
	destroy           func() error
	listFilesError    error
	readFileError     error
	panic             bool
	root              *memFilesystemNode
	statFileError     error
}

func (fs *memFilesystem) AbsolutePath(path string) (string, error) {
	if fs.absolutePathError != nil {
		if fs.panic {
			panic(fs.absolutePathError)
		}

		return "", fs.absolutePathError
	}

	return fs.FilesystemDefaults.AbsolutePath(path)
}

func (fs *memFilesystem) Destroy() error {
	if fs.destroy != nil {
		return fs.destroy()
	}

	return nil
}

func (fs *memFilesystem) ListFiles(path string) ([]os.FileInfo, error) {
	var children []os.FileInfo
	var node *memFilesystemNode

	if fs.listFilesError != nil {
		if fs.panic {
			panic(fs.listFilesError)
		}

		return nil, fs.listFilesError
	}

	node = fs.findNode(path)

	if node == nil {
		return nil, memFilesystemErrorNotFound
	}

	if !node.IsDir() {
		return []os.FileInfo{node}, nil
	}

	for name, child := range node.children {
		// We don't want to redundantly provide node names when specifying them as literals, so we'll take this
		// opportunity to assign the correct name here.

		child.name = name

		children = append(children, child)
	}

	return children, nil
}

func (fs *memFilesystem) ReadFile(path string) (io.ReadCloser, error) {
	var node *memFilesystemNode
	var reader io.Reader

	if fs.readFileError != nil {
		if fs.panic {
			panic(fs.readFileError)
		}

		return nil, fs.readFileError
	}

	node = fs.findNode(path)

	if node == nil {
		return nil, memFilesystemErrorNotFound
	}

	reader = bytes.NewBufferString(node.contents)

	return ioutil.NopCloser(reader), nil
}

func (fs *memFilesystem) StatFile(path string) (os.FileInfo, error) {
	var node *memFilesystemNode

	if fs.statFileError != nil {
		if fs.panic {
			panic(fs.statFileError)
		}

		return nil, fs.statFileError
	}

	node = fs.findNode(path)

	if node == nil {
		return nil, memFilesystemErrorNotFound
	}

	return node, nil
}

func (fs *memFilesystem) findNode(path string) *memFilesystemNode {
	var curNode = fs.root
	var split []string

	if path == "" || path == "/" {
		return fs.root
	}

	split = strings.Split(path, "/")

	if split[0] == "" {
		split = split[1:]
	}

	for _, segment := range split {
		curNode = curNode.children[segment]

		if curNode == nil {
			return nil
		} else {
			// So we don't have to redundantly specify the name when we're creating memFilesystemNode literals in our
			// testcases.

			curNode.name = segment
		}
	}

	return curNode
}

// In-memory Filesystem node.
type memFilesystemNode struct {
	children map[string]*memFilesystemNode
	contents string
	name     string
}

func (node *memFilesystemNode) Name() string {
	return node.name
}

func (node *memFilesystemNode) Size() int64 {
	return 0
}

func (node *memFilesystemNode) Mode() os.FileMode {
	return os.ModePerm
}

func (node *memFilesystemNode) ModTime() time.Time {
	return time.Now()
}

func (node *memFilesystemNode) IsDir() bool {
	return len(node.children) > 0
}

func (node *memFilesystemNode) Sys() interface{} {
	return nil
}

// Dummy os.FileInfo implementation used to test file and fileInfoStack.
type nilFileInfo struct {
	name string
}

func (fi *nilFileInfo) Name() string {
	return fi.name
}

func (fi *nilFileInfo) Size() int64 {
	return 0
}

func (fi *nilFileInfo) Mode() os.FileMode {
	return os.ModePerm
}

func (fi *nilFileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi *nilFileInfo) IsDir() bool {
	return false
}

func (fi *nilFileInfo) Sys() interface{} {
	return nil
}

type wrappedError interface {
	error

	Unwrap() error
}

// EventSink implementation used to verify the order of sent events.
type testEventSink struct {
	events []Event
}

func (sink *testEventSink) Send(event Event) {
	sink.events = append(sink.events, event)
}

func (sink *testEventSink) expectEvents(events ...string) {
	// Ugly, but we have to sync against the globalEventSink if we want to introspect on it during multithreaded
	// testing.

	globalEventSink.mutex.RLock()

	So(sink.events, ShouldHaveLength, len(events))

	for i, event := range sink.events {
		var index = strings.Index(events[i], ".")

		So(event.Component(), ShouldEqual, events[i][:index])
		So(event.Type(), ShouldEqual, events[i][index+1:])
	}

	globalEventSink.mutex.RUnlock()
}

//
// Private constants
//

const (
	// Filter event names for testEventSink.expectEvents()
	eventFilterCancelled      = componentFilter + "." + eventTypeCancelled
	eventFilterCreated        = componentFilter + "." + eventTypeCreated
	eventFilterDestroyed      = componentFilter + "." + eventTypeDestroyed
	eventFilterFinished       = componentFilter + "." + eventTypeFinished
	eventFilterResultProduced = componentFilter + "." + eventTypeResultProduced
	eventFilterStarted        = componentFilter + "." + eventTypeStarted

	// Source event names for testEventSink.expectEvents()
	eventSourceCancelled      = componentSource + "." + eventTypeCancelled
	eventSourceCreated        = componentSource + "." + eventTypeCreated
	eventSourceDestroyed      = componentSource + "." + eventTypeDestroyed
	eventSourceFinished       = componentSource + "." + eventTypeFinished
	eventSourceResultProduced = componentSource + "." + eventTypeResultProduced
	eventSourceStarted        = componentSource + "." + eventTypeStarted
)

//
// Private variables
//

var (
	eventSinkTestMutex sync.Mutex

	idsInvalid = []string{"", " ", ".", "a ", " a", "a.", ".a", "a..b", "a-b", "?"}
	idsValid   = []string{"a", "0", "a.0", "0.1", "a.b.c", "abc.def", "0.1.2", "0.abc.1.def"}

	memFilesystemErrorNotFound = fmt.Errorf("path not found")
)

//
// Private functions
//

func collectSourceResults(source Source) []Result {
	var in <-chan Result
	var results = make([]Result, 0)

	in, _ = source.Files(NewContext(ContextConfig{}))

	for result := range in {
		results = append(results, result)
	}

	return results
}

func expectFilePathsInResults(c C, results []Result, paths []string) {
	var pathMap = make(map[string]bool)

	for _, path := range paths {
		pathMap[path] = true
	}

	if c == nil {
		for _, result := range results {
			So(result.Error(), ShouldBeNil)
			So(pathMap, ShouldContainKey, result.File().Path().String())
		}
	} else {
		for _, result := range results {
			c.So(result.Error(), ShouldBeNil)
			c.So(pathMap, ShouldContainKey, result.File().Path().String())
		}
	}
}

func resetGlobalEventSink() {
	globalEventSink.mutex.Lock()

	globalEventSink.allowedMap = make(map[string]bool)
	globalEventSink.children = make([]EventSink, 0)

	globalEventSink.mutex.Unlock()
}

func testEventMarshalUnmarshal(evt Event) Event {
	var contents []byte
	var err error
	var unmarshalled = new(event)
	contents, err = json.Marshal(evt)

	So(err, ShouldBeNil)
	So(contents, ShouldNotBeNil)

	err = json.Unmarshal(contents, unmarshalled)

	So(err, ShouldBeNil)

	So(unmarshalled.Component(), ShouldEqual, evt.Component())
	So(unmarshalled.Data(), ShouldResemble, evt.Data())
	So(unmarshalled.Type(), ShouldEqual, evt.Type())

	return unmarshalled
}

func validateSourceEvent(event Event, component, eventType, id string) {
	Convey("it should create the correct event", func() {
		So(event.Component(), ShouldEqual, component)
		So(event.Data(), ShouldContainKey, eventFieldID)
		So(event.Data()[eventFieldID], ShouldEqual, id)

		if eventType == eventTypeResultProduced {
			So((event.Data()[eventFieldError] != nil) || (event.Data()[eventFieldFile] != nil), ShouldBeTrue)
		}

		So(event.Type(), ShouldEqual, eventType)
	})

	Convey("it should be possible to marshal the event to JSON and unmarshal it", func() {
		testEventMarshalUnmarshal(event)
	})
}
