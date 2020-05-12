package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"encoding/json"
	"errors"
	"sync"

	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Testcases
//

// Utility function tests

var _ = g.Describe("allowEventsFrom", func() {
	g.Describe("calling allowEventsFrom", func() {
		g.Context("with an empty string", func() {
			g.It("should be ignored", func() {
				var size int

				globalEventSink.mutex.RLock()

				size = len(globalEventSink.allowedMap)

				globalEventSink.mutex.RUnlock()

				allowEventsFrom("", true)
				allowEventsFrom(" ", true)

				globalEventSink.mutex.RLock()
				defer globalEventSink.mutex.RUnlock()

				Expect(len(globalEventSink.allowedMap)).To(Equal(size))
			})
		})

		g.Context("with a valid component", func() {
			g.It("should allow or disallow events properly", func() {
				var sink = newTestEventSink()

				RegisterEventSink(sink)

				// Kind of ugly, but we have to do this so as not to interfere with other tests.

				globalEventSink.mutex.Lock()
				globalEventSink.mutex.Unlock()

				allowEventsFromInternal(componentSource, false)

				globalEventSink.sendInternal(sourceEventCreated(sink.id))
				globalEventSink.sendInternal(sourceEventDestroyed(sink.id))

				Expect(sink.events).To(HaveLen(0))

				allowEventsFromInternal(componentSource, true)

				globalEventSink.sendInternal(sourceEventCreated(sink.id))
				globalEventSink.sendInternal(sourceEventDestroyed(sink.id))

				Expect(sink).To(haveTheseEvents(eventSourceCreated, eventSourceDestroyed))
			})
		})
	})
})

var _ = g.Describe("newEvent", func() {
	g.Describe("given a new instance", func() {
		var event Event

		g.BeforeEach(func() {
			event = newEvent(componentSource, "source", eventTypeCreated)
		})

		g.Describe("calling Component", func() {
			g.It("should return the expected component", func() {
				Expect(event.Component()).To(Equal(componentSource))
			})
		})

		g.Describe("calling Data", func() {
			g.It("should return the expected data", func() {
				Expect(event.Data()).NotTo(BeNil())
				Expect(event.Data()).To(HaveLen(1))
				Expect(event.Data()).To(HaveKey(eventFieldID))
				Expect(event.Data()[eventFieldID]).To(Equal("source"))
			})
		})

		g.Describe("calling Type", func() {
			g.It("should return the expected type", func() {
				Expect(event.Type()).To(Equal(eventTypeCreated))
			})
		})
	})
})

var _ = g.Describe("newResultProducedEvent", func() {
	g.Describe("calling newResultProducedEvent", func() {
		g.Specify("should return an Event that can be marshalled to and unmarshalled from JSON", func() {
			var contents []byte
			var err error
			var evt Event
			var res = &result{
				err: errors.New("result error"),
				file: &file{
					fileInfo: &nilFileInfo{
						name: "name",
					},
					path: newFilePath(nil, "name", "/"),
				},
			}
			var unmarshalled = new(event)

			evt = newResultProducedEvent(componentSource, "source", res)

			contents, err = json.Marshal(evt)

			Expect(err).To(BeNil())
			Expect(contents).NotTo(BeNil())

			err = json.Unmarshal(contents, unmarshalled)

			Expect(err).To(BeNil())

			Expect(unmarshalled.Component()).To(Equal(evt.Component()))
			Expect(unmarshalled.Data()).NotTo(BeNil())
			Expect(unmarshalled.Data()).To(HaveKey(eventFieldError))
			Expect(unmarshalled.Data()[eventFieldError]).To(Equal("result error"))
			Expect(unmarshalled.Data()).To(HaveKey(eventFieldFile))
			Expect(unmarshalled.Data()[eventFieldFile]).To(Equal(res.File().Path().String()))
			Expect(unmarshalled.Type()).To(Equal(evt.Type()))
		})
	})
})

var _ = g.Describe("RegisterEventSink", func() {
	g.Describe("calling RegisterEventSink", func() {
		g.BeforeEach(globalEventSink.mutex.Lock)
		g.AfterEach(globalEventSink.mutex.Unlock)

		g.Context("with a nil EventSink", func() {
			g.It("should be ignored", func() {
				var size = len(globalEventSink.children)

				RegisterEventSink(nil)

				Expect(len(globalEventSink.children)).To(Equal(size))
			})
		})
	})
})

var _ = g.Describe("sendEvent", func() {
	g.Describe("calling sendEvent", func() {
		var sink *testEventSink

		g.BeforeEach(func() {
			sink = newTestEventSink()

			RegisterEventSink(sink)
		})

		g.Context("with a nil Event", func() {
			g.It("should be ignored", func() {
				Expect(sink.events).To(HaveLen(0))

				sendEvent(nil)

				Expect(sink.events).To(HaveLen(0))
			})
		})
	})
})

//
// Private types
//

// EventSink implementation used to verify the order of sent events.
type testEventSink struct {
	events []Event
	id     string
	mutex  sync.Mutex
}

func (sink *testEventSink) Send(event Event) {
	if id, ok := event.Data()[eventFieldID]; ok {
		if id != sink.id {
			return
		}
	} else {
		return
	}

	sink.mutex.Lock()

	sink.events = append(sink.events, event)

	sink.mutex.Unlock()
}
