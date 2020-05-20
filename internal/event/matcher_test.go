package event // import "golang.handcraftedbits.com/pipewerx/internal/event"

import (
	"errors"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

//
// Private types
//

// types.GomegaMatcher implementation used to ensure that an EventSink contains the given events in the exact order
// specified.
type matcherHaveTheseEvents struct {
	events []string
}

func (matcher *matcherHaveTheseEvents) Match(actual interface{}) (bool, error) {
	var ok bool
	var sink *testSink

	sink, ok = actual.(*testSink)

	if !ok {
		return false, errors.New("haveTheseEvents expects *event.testSink")
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

func haveTheseEvents(events ...string) types.GomegaMatcher {
	return &matcherHaveTheseEvents{
		events: events,
	}
}
