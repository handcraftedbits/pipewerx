package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"strings"
	"sync"
)

//
// Public types
//

type Event interface {
	Component() string

	Data() map[string]interface{}

	Type() string
}

type EventSink interface {
	Send(event Event)
}

//
// Public functions
//

func RegisterEventSink(eventSink EventSink) {
	if eventSink == nil {
		return
	}

	globalEventSink.mutex.Lock()

	globalEventSink.children = append(globalEventSink.children, eventSink)

	globalEventSink.mutex.Unlock()
}

//
// Private types
//

// EventSink implementation that delegates to child EventSinks.
type delegatingEventSink struct {
	allowedMap map[string]bool
	children   []EventSink
	mutex      sync.RWMutex
}

func (sink *delegatingEventSink) Send(event Event) {
	if event == nil {
		return
	}

	sink.mutex.Lock()
	defer sink.mutex.Unlock()

	for _, child := range sink.children {
		if globalEventSink.allowedMap[event.Component()] {
			child.Send(event)
		}
	}
}

// Event implementation
type event struct {
	EventComponent string                 `json:"component"`
	EventData      map[string]interface{} `json:"data"`
	EventType      string                 `json:"type"`
}

func (e *event) Component() string {
	return e.EventComponent
}

func (e *event) Data() map[string]interface{} {
	return e.EventData
}

func (e *event) Type() string {
	return e.EventType
}

//
// Private constants
//

const (
	eventFieldError = "error"
	eventFieldFile  = "file"
	eventFieldID    = "id"

	eventTypeCancelled      = "cancelled"
	eventTypeCreated        = "created"
	eventTypeDestroyed      = "destroyed"
	eventTypeFinished       = "finished"
	eventTypeResultProduced = "resultProduced"
	eventTypeStarted        = "started"
)

//
// Private variables
//

var (
	globalEventSink = &delegatingEventSink{
		allowedMap: make(map[string]bool),
		mutex:      sync.RWMutex{},
	}
)

//
// Private functions
//

func allowEventsFrom(component string, shouldAllow bool) {
	if strings.TrimSpace(component) == "" {
		return
	}

	globalEventSink.mutex.Lock()

	globalEventSink.allowedMap[component] = shouldAllow

	globalEventSink.mutex.Unlock()
}

func isEventAllowedFrom(component string) bool {
	var result bool

	globalEventSink.mutex.RLock()

	result = globalEventSink.allowedMap[component]

	globalEventSink.mutex.RUnlock()

	return result
}

func newEvent(component, id, eventType string) Event {
	return &event{
		EventComponent: component,
		EventData: map[string]interface{}{
			eventFieldID: id,
		},
		EventType: eventType,
	}
}

func newResultProducedEvent(component, id string, result Result) Event {
	var event = newEvent(component, id, eventTypeResultProduced)

	if result.Error() != nil {
		event.Data()[eventFieldError] = result.Error().Error()
	}

	if result.File() != nil {
		event.Data()[eventFieldFile] = result.File().Path().String()
	}

	return event
}

func sendEvent(event Event) {
	globalEventSink.Send(event)
}
