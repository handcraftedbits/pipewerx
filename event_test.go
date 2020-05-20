package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.handcraftedbits.com/pipewerx/internal/event"
)

//
// Testcases
//

// File event tests

var _ = g.Describe("File events", func() {
	var f = &file{
		fileInfo: &nilFileInfo{
			name: "name.ext",
			size: 10,
		},
		path: newFilePath(nil, "name.ext", "/"),
	}
	var sourceID = "source"

	g.Describe("calling fileEventClosed", func() {
		g.It("should return a valid event", func() {
			Expect(fileEventClosed(f, sourceID)).To(beAValidFileEvent(event.TypeClosed, sourceID, f.Path().String(), 0))
		})
	})

	g.Describe("calling fileEventOpened", func() {
		g.It("should return a valid event", func() {
			Expect(fileEventOpened(f, sourceID)).To(beAValidFileEvent(event.TypeOpened, sourceID, f.Path().String(), 10))
		})
	})

	g.Describe("calling fileEventRead", func() {
		g.It("should return a valid event", func() {
			Expect(fileEventRead(f, sourceID, 5)).To(beAValidFileEvent(event.TypeRead, sourceID, f.Path().String(), 5))
		})
	})
})

// Filter event tests

var _ = g.Describe("Filter events", func() {
	var id = "filter"

	g.Describe("calling filterEventCancelled", func() {
		g.It("should return a valid event", func() {
			Expect(filterEventCancelled(id)).To(beAValidEvent(componentFilter, event.TypeCancelled, id))
		})
	})

	g.Describe("calling filterEventCreated", func() {
		g.It("should return a valid event", func() {
			Expect(filterEventCreated(id)).To(beAValidEvent(componentFilter, event.TypeCreated, id))
		})
	})

	g.Describe("calling filterEventDestroyed", func() {
		g.It("should return a valid event", func() {
			Expect(filterEventDestroyed(id)).To(beAValidEvent(componentFilter, event.TypeDestroyed, id))
		})
	})

	g.Describe("calling filterEventFinished", func() {
		g.It("should return a valid event", func() {
			Expect(filterEventFinished(id)).To(beAValidEvent(componentFilter, event.TypeFinished, id))
		})
	})

	g.Describe("calling filterEventResultProduced", func() {
		g.It("should return a valid event", func() {
			var res = &result{
				err: errors.New("result error"),
				file: &file{
					fileInfo: &nilFileInfo{
						name: "name",
					},
					path: newFilePath(nil, "name", "/"),
				},
			}

			Expect(filterEventResultProduced(id, res)).To(beAValidEvent(componentFilter, event.TypeResultProduced, id))
		})
	})

	g.Describe("calling filterEventStarted", func() {
		g.It("should return a valid event", func() {
			Expect(filterEventStarted(id)).To(beAValidEvent(componentFilter, event.TypeStarted, id))
		})
	})
})

// Source event tests

var _ = g.Describe("Source events", func() {
	var id = "source"

	g.Describe("calling sourceEventCancelled", func() {
		g.It("should return a valid event", func() {
			Expect(sourceEventCancelled(id)).To(beAValidEvent(componentSource, event.TypeCancelled, id))
		})
	})

	g.Describe("calling sourceEventCreated", func() {
		g.It("should return a valid event", func() {
			Expect(sourceEventCreated(id)).To(beAValidEvent(componentSource, event.TypeCreated, id))
		})
	})

	g.Describe("calling sourceEventDestroyed", func() {
		g.It("should return a valid event", func() {
			Expect(sourceEventDestroyed(id)).To(beAValidEvent(componentSource, event.TypeDestroyed, id))
		})
	})

	g.Describe("calling sourceEventFinished", func() {
		g.It("should return a valid event", func() {
			Expect(sourceEventFinished(id)).To(beAValidEvent(componentSource, event.TypeFinished, id))
		})
	})

	g.Describe("calling sourceEventResultProduced", func() {
		g.It("should return a valid event", func() {
			var res = &result{
				err: errors.New("result error"),
				file: &file{
					fileInfo: &nilFileInfo{
						name: "name",
					},
					path: newFilePath(nil, "name", "/"),
				},
			}

			Expect(sourceEventResultProduced(id, res)).To(beAValidEvent(componentSource, event.TypeResultProduced, id))
		})
	})

	g.Describe("calling sourceEventStarted", func() {
		g.It("should return a valid event", func() {
			Expect(sourceEventStarted(id)).To(beAValidEvent(componentSource, event.TypeStarted, id))
		})
	})
})

// Utility function tests

var _ = g.Describe("newResultProducedEvent", func() {
	g.Describe("calling newResultProducedEvent", func() {
		g.Specify("should return an Event that can be marshalled to and unmarshalled from JSON", func() {
			var contents []byte
			var err error
			var evt event.Event
			var res = &result{
				err: errors.New("result error"),
				file: &file{
					fileInfo: &nilFileInfo{
						name: "name",
					},
					path: newFilePath(nil, "name", "/"),
				},
			}
			var unmarshalled = event.WithID("", "", "")

			evt = newResultProducedEvent(componentSource, "source", res)

			contents, err = json.Marshal(evt)

			Expect(err).To(BeNil())
			Expect(contents).NotTo(BeNil())

			err = json.Unmarshal(contents, unmarshalled)

			Expect(err).To(BeNil())

			Expect(unmarshalled.Component()).To(Equal(evt.Component()))
			Expect(unmarshalled.Data()).NotTo(BeNil())
			Expect(unmarshalled.Data()).To(HaveKey(event.FieldError))
			Expect(unmarshalled.Data()[event.FieldError]).To(Equal("result error"))
			Expect(unmarshalled.Data()).To(HaveKey(event.FieldFile))
			Expect(unmarshalled.Data()[event.FieldFile]).To(Equal(res.File().Path().String()))
			Expect(unmarshalled.Type()).To(Equal(evt.Type()))
		})
	})
})

//
// Private types
//

// Sink implementation used to verify the order of sent events.
type testEventSink struct {
	events []event.Event
	id     string
	mutex  sync.Mutex
}

func (sink *testEventSink) Send(evt event.Event) {
	if id, ok := evt.Data()[event.FieldID]; ok {
		if id != sink.id {
			return
		}
	} else {
		return
	}

	sink.mutex.Lock()

	sink.events = append(sink.events, evt)

	sink.mutex.Unlock()
}

//
// Private functions
//

func newTestEventSink() *testEventSink {
	var id = fmt.Sprintf("%d", rand.Int())

	return &testEventSink{
		id:    id,
		mutex: sync.Mutex{},
	}
}
