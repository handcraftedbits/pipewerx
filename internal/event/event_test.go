package event // import "golanhandcraftedbits.com/pipewerx/internal/event"

import (
	"fmt"
	"math/rand"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Testcases
//

// event tests

var _ = Describe("event", func() {
	Describe("given a new instance", func() {
		var component = "withIDTest"
		var event Event

		BeforeEach(func() {
			event = WithID(component, "source", TypeCreated)
		})

		Describe("calling Component", func() {
			It("should return the expected component", func() {
				Expect(event.Component()).To(Equal(component))
			})
		})

		Describe("calling Data", func() {
			It("should return the expected data", func() {
				Expect(event.Data()).NotTo(BeNil())
				Expect(event.Data()).To(HaveLen(1))
				Expect(event.Data()).To(HaveKey(FieldID))
				Expect(event.Data()[FieldID]).To(Equal("source"))
			})
		})

		Describe("calling Type", func() {
			It("should return the expected type", func() {
				Expect(event.Type()).To(Equal(TypeCreated))
			})
		})
	})
})

// Utility function tests

var _ = Describe("AllowFrom", func() {
	Describe("calling AllowFrom", func() {
		Context("with an empty string", func() {
			It("should be ignored", func() {
				var size int

				globalSink.mutex.RLock()

				size = len(globalSink.allowedMap)

				globalSink.mutex.RUnlock()

				AllowFrom("", true)
				AllowFrom(" ", true)

				globalSink.mutex.RLock()
				defer globalSink.mutex.RUnlock()

				Expect(len(globalSink.allowedMap)).To(Equal(size))
			})
		})

		Context("with a valid component", func() {
			It("should allow or disallow events properly", func() {
				var component = "allowTest"
				var sink = newTestSink()

				RegisterSink(sink)

				AllowFrom(component, false)

				Send(WithID(component, sink.id, TypeCreated))
				Send(WithID(component, sink.id, TypeDestroyed))

				Expect(sink.events).To(HaveLen(0))

				AllowFrom(component, true)

				Send(WithID(component, sink.id, TypeCreated))
				Send(WithID(component, sink.id, TypeDestroyed))

				Expect(sink).To(haveTheseEvents(component+"."+TypeCreated, component+"."+TypeDestroyed))
			})
		})
	})
})

var _ = Describe("IsAllowedFrom", func() {
	Describe("calling IsAllowedFrom", func() {
		Context("with a valid component", func() {
			It("should return the correct values", func() {
				var component = "isAllowedFromTest"

				AllowFrom(component, false)

				Expect(IsAllowedFrom(component)).To(BeFalse())

				AllowFrom(component, true)

				Expect(IsAllowedFrom(component)).To(BeTrue())
			})
		})
	})
})

var _ = Describe("RegisterSink", func() {
	Describe("calling RegisterSink", func() {
		BeforeEach(globalSink.mutex.Lock)
		AfterEach(globalSink.mutex.Unlock)

		Context("with a nil Sink", func() {
			It("should be ignored", func() {
				var size = len(globalSink.children)

				RegisterSink(nil)

				Expect(len(globalSink.children)).To(Equal(size))
			})
		})
	})
})

var _ = Describe("Send", func() {
	Describe("calling Send", func() {
		var sink *testSink

		BeforeEach(func() {
			sink = newTestSink()

			RegisterSink(sink)
		})

		Context("with a nil Event", func() {
			It("should be ignored", func() {
				Expect(sink.events).To(HaveLen(0))

				Send(nil)

				Expect(sink.events).To(HaveLen(0))
			})
		})
	})
})

//
// Private types
//

// Sink implementation used to verify the order of sent events.
type testSink struct {
	events []Event
	id     string
	mutex  sync.Mutex
}

func (sink *testSink) Send(event Event) {
	if id, ok := event.Data()[FieldID]; ok {
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

//
// Private functions
//

func newTestSink() *testSink {
	var id = fmt.Sprintf("%d", rand.Int())

	return &testSink{
		id:    id,
		mutex: sync.Mutex{},
	}
}
