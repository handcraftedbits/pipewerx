package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"encoding/json"
	"errors"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

//
// Private types
//

// types.GomegaMatcher implementation used to ensure that an Event is properly constructed and can be marshalled to and
// unmarshalled from JSON.
type matcherBeAValidEvent struct {
	component string
	eventType string
	id        string
}

func (matcher *matcherBeAValidEvent) Match(actual interface{}) (bool, error) {
	var contents []byte
	var err error
	var evt Event
	var ok bool
	var unmarshalled = new(event)

	evt, ok = actual.(Event)

	if !ok {
		return false, errors.New("beAValidEvent expected pipewerx.Event")
	}

	Expect(evt.Component()).To(Equal(matcher.component))
	Expect(evt.Data()).NotTo(BeNil())
	Expect(evt.Data()).To(HaveKey(eventFieldID))
	Expect(evt.Data()[eventFieldID]).To(Equal(matcher.id))
	Expect(evt.Type()).To(Equal(matcher.eventType))

	if matcher.eventType == eventTypeResultProduced {
		Expect(evt.Data()).To(SatisfyAny(
			HaveKey(eventFieldError),
			HaveKey(eventFieldFile)))
	}

	contents, err = json.Marshal(evt)

	Expect(err).To(BeNil())
	Expect(contents).NotTo(BeNil())

	err = json.Unmarshal(contents, unmarshalled)

	Expect(err).To(BeNil())

	Expect(unmarshalled.Component()).To(Equal(evt.Component()))
	Expect(unmarshalled.Data()).To(Equal(evt.Data()))
	Expect(unmarshalled.Type()).To(Equal(evt.Type()))

	return true, nil
}

func (matcher *matcherBeAValidEvent) FailureMessage(actual interface{}) string {
	return ""
}

func (matcher *matcherBeAValidEvent) NegatedFailureMessage(actual interface{}) string {
	return ""
}

// types.GomegaMatcher implementation used to ensure that a File Event is properly constructed and can be marshalled to
// and unmarshalled from JSON.
type matcherBeAValidFileEvent struct {
	eventType      string
	length         int64
	path           string
	sourceOrDestID string
}

func (matcher *matcherBeAValidFileEvent) Match(actual interface{}) (bool, error) {
	var contents []byte
	var err error
	var evt Event
	var ok bool
	var unmarshalled = new(event)

	evt, ok = actual.(Event)

	if !ok {
		return false, errors.New("beAValidFileEvent expected pipewerx.Event")
	}

	Expect(evt.Component()).To(Equal(componentFile))
	Expect(evt.Data()).NotTo(BeNil())
	Expect(evt.Data()).To(HaveKey(eventFieldFile))
	Expect(evt.Data()[eventFieldFile]).To(Equal(matcher.path))
	Expect(evt.Data()).To(HaveKey(eventFieldID))
	Expect(evt.Data()[eventFieldID]).To(Equal(matcher.sourceOrDestID))
	Expect(evt.Type()).To(Equal(matcher.eventType))

	if matcher.eventType == eventTypeOpened || matcher.eventType == eventTypeRead {
		Expect(evt.Data()).To(HaveKey(eventFieldLength))
		Expect(evt.Data()[eventFieldLength]).To(BeEquivalentTo(matcher.length))

		// Kind of annoying, but when the JSON is unmarshalled the length field will be float64, and reflect.DeepEqual()
		// is doing strict comparisons; so unless we cast the "expected" map with the correct type, everything will
		// fail.

		if matcher.eventType == eventTypeRead {
			evt.Data()[eventFieldLength] = float64(evt.Data()[eventFieldLength].(int))
		} else {
			evt.Data()[eventFieldLength] = float64(evt.Data()[eventFieldLength].(int64))
		}
	}

	contents, err = json.Marshal(evt)

	Expect(err).To(BeNil())
	Expect(contents).NotTo(BeNil())

	err = json.Unmarshal(contents, unmarshalled)

	Expect(err).To(BeNil())

	Expect(unmarshalled.Component()).To(Equal(evt.Component()))
	Expect(unmarshalled.Data()).To(Equal(evt.Data()))
	Expect(unmarshalled.Type()).To(Equal(evt.Type()))

	return true, nil
}

func (matcher *matcherBeAValidFileEvent) FailureMessage(actual interface{}) string {
	return ""
}

func (matcher *matcherBeAValidFileEvent) NegatedFailureMessage(actual interface{}) string {
	return ""
}

// types.GomegaMatcher implementation used to ensure that an array of Result objects contains all or some FilePaths from
// a given set.
type matcherHaveAllOrSomeOfTheseFilePaths struct {
	paths []string
}

func (matcher *matcherHaveAllOrSomeOfTheseFilePaths) Match(actual interface{}) (bool, error) {
	var ok bool
	var pathMap map[string]bool
	var results []Result

	results, ok = actual.([]Result)

	if !ok {
		return false, errors.New("haveAllOrSomeOfTheseFilePaths expects []pipewerx.Result")
	}

	pathMap = make(map[string]bool)

	for _, path := range matcher.paths {
		pathMap[path] = true
	}

	for _, result := range results {
		Expect(result.Error()).To(BeNil())
		Expect(pathMap).To(HaveKey(result.File().Path().String()))
	}

	return true, nil
}

func (matcher *matcherHaveAllOrSomeOfTheseFilePaths) FailureMessage(actual interface{}) string {
	return ""
}

func (matcher *matcherHaveAllOrSomeOfTheseFilePaths) NegatedFailureMessage(actual interface{}) string {
	return ""
}

// types.GomegaMatcher implementation used to ensure that an EventSink contains the given events in the exact order
// specified.
type matcherHaveTheseEvents struct {
	events []string
}

func (matcher *matcherHaveTheseEvents) Match(actual interface{}) (bool, error) {
	var ok bool
	var sink *testEventSink

	sink, ok = actual.(*testEventSink)

	if !ok {
		return false, errors.New("haveTheseEvents expects *pipewerx.testEventSink")
	}

	sink.mutex.Lock()
	defer sink.mutex.Unlock()

	Expect(sink.events).To(HaveLen(len(matcher.events)))

	for i, event := range sink.events {
		var index = strings.Index(matcher.events[i], ".")

		Expect(event.Component()).To(Equal(matcher.events[i][:index]))
		Expect(event.Type()).To(Equal(matcher.events[i][index+1:]))
	}

	return true, nil
}

func (matcher *matcherHaveTheseEvents) FailureMessage(actual interface{}) string {
	return ""
}

func (matcher *matcherHaveTheseEvents) NegatedFailureMessage(actual interface{}) string {
	return ""
}

//
// Private functions
//

func beAValidEvent(component, eventType, id string) types.GomegaMatcher {
	return &matcherBeAValidEvent{
		component: component,
		eventType: eventType,
		id:        id,
	}
}

func beAValidFileEvent(eventType, sourceID, path string, length int64) types.GomegaMatcher {
	return &matcherBeAValidFileEvent{
		eventType:      eventType,
		length:         length,
		path:           path,
		sourceOrDestID: sourceID,
	}
}

func collectSourceResults(source Source) []Result {
	var in <-chan Result
	var results = make([]Result, 0)

	in, _ = source.Files(NewContext(ContextConfig{}))

	for result := range in {
		results = append(results, result)
	}

	return results
}

func haveAllOrSomeOfTheseFilePaths(paths ...string) types.GomegaMatcher {
	return &matcherHaveAllOrSomeOfTheseFilePaths{
		paths: paths,
	}
}

func haveTheseEvents(events ...string) types.GomegaMatcher {
	return &matcherHaveTheseEvents{
		events: events,
	}
}
