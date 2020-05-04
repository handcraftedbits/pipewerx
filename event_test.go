package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// Utility function tests

func TestAllowEventsFrom(t *testing.T) {
	eventSinkTestMutex.Lock()
	defer eventSinkTestMutex.Unlock()

	Convey("When calling allowEventsFrom", t, func() {
		Reset(resetGlobalEventSink)

		Convey("it should ignore an empty string", func() {
			So(globalEventSink.allowedMap, ShouldHaveLength, 0)

			allowEventsFrom("", true)
			allowEventsFrom(" ", true)

			So(globalEventSink.allowedMap, ShouldHaveLength, 0)
		})

		Convey("it should allow or disallow events from a particular component", func() {
			var sink = &testEventSink{}

			RegisterEventSink(sink)

			sendEvent(sourceEventCreated("source"))
			sendEvent(sourceEventDestroyed("source"))

			So(sink.events, ShouldHaveLength, 0)

			allowEventsFrom(componentSource, true)

			sendEvent(sourceEventCreated("source"))
			sendEvent(sourceEventDestroyed("source"))

			sink.expectEvents(eventSourceCreated, eventSourceDestroyed)

			allowEventsFrom(componentSource, false)

			sendEvent(sourceEventCreated("source"))
			sendEvent(sourceEventDestroyed("source"))

			sink.expectEvents(eventSourceCreated, eventSourceDestroyed)
		})
	})
}

func TestNewEvent(t *testing.T) {
	Convey("When calling newEvent", t, func() {
		var event = newEvent(componentSource, "source", eventTypeCreated)

		Convey("calling Component should return the expected component", func() {
			So(event.Component(), ShouldEqual, componentSource)
		})

		Convey("calling Data should return the expected data", func() {
			So(event.Data(), ShouldHaveLength, 1)
			So(event.Data(), ShouldContainKey, eventFieldID)
			So(event.Data()[eventFieldID], ShouldEqual, "source")
		})

		Convey("calling Type should return the expected type", func() {
			So(event.Type(), ShouldEqual, eventTypeCreated)
		})

		Convey("it should be possible to marshal the event to JSON and unmarshal it", func() {
			testEventMarshalUnmarshal(event)
		})
	})
}

func TestRegisterEventSink(t *testing.T) {
	eventSinkTestMutex.Lock()
	defer eventSinkTestMutex.Unlock()

	Convey("When calling RegisterEventSink", t, func() {
		Reset(resetGlobalEventSink)

		Convey("it should ignore a nil EventSink", func() {
			So(globalEventSink.children, ShouldHaveLength, 0)

			RegisterEventSink(nil)

			So(globalEventSink.children, ShouldHaveLength, 0)
		})

		Convey("it should successfully register a non-nil EventSink", func() {
			So(globalEventSink.children, ShouldHaveLength, 0)

			RegisterEventSink(&testEventSink{})

			So(globalEventSink.children, ShouldHaveLength, 1)
		})
	})
}

func TestNewResultProducedEvent(t *testing.T) {
	Convey("When calling newResultProducedEvent", t, func() {
		var event Event
		var res = &result{
			err: errors.New("result error"),
			file: &file{
				fileInfo: &nilFileInfo{
					name: "name",
				},
				path: newFilePath(nil, "name", "/"),
			},
		}
		var unmarshalled Event

		event = newResultProducedEvent(componentSource, "source", res)
		unmarshalled = testEventMarshalUnmarshal(event)

		So(unmarshalled.Data(), ShouldContainKey, eventFieldError)
		So(unmarshalled.Data()[eventFieldError], ShouldEqual, "result error")
		So(unmarshalled.Data(), ShouldContainKey, eventFieldFile)
		So(unmarshalled.Data()[eventFieldFile], ShouldEqual, res.File().Path().String())
	})
}

func TestSendEvent(t *testing.T) {
	eventSinkTestMutex.Lock()
	defer eventSinkTestMutex.Unlock()

	Convey("When calling sendEvent", t, func() {
		Reset(resetGlobalEventSink)

		Convey("it should ignore a nil Event", func() {
			var sink = &testEventSink{}

			RegisterEventSink(sink)

			So(sink.events, ShouldHaveLength, 0)

			sendEvent(nil)

			So(sink.events, ShouldHaveLength, 0)
		})
	})
}
