package event // import "golang.handcraftedbits.com/pipewerx/internal/event"

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

type Sink interface {
	Send(event Event)
}

//
// Public constants
//

const (
	FieldError  = "error"
	FieldFile   = "file"
	FieldID     = "id"
	FieldLength = "length"

	TypeCancelled      = "cancelled"
	TypeClosed         = "closed"
	TypeCreated        = "created"
	TypeDestroyed      = "destroyed"
	TypeFinished       = "finished"
	TypeOpened         = "opened"
	TypeRead           = "read"
	TypeResultProduced = "resultProduced"
	TypeStarted        = "started"
)

//
// Public functions
//

func AllowFrom(component string, shouldAllow bool) {
	if strings.TrimSpace(component) == "" {
		return
	}

	globalSink.mutex.Lock()

	allowFromInternal(component, shouldAllow)

	globalSink.mutex.Unlock()
}

func IsAllowedFrom(component string) bool {
	var result bool

	globalSink.mutex.RLock()

	result = globalSink.allowedMap[component]

	globalSink.mutex.RUnlock()

	return result
}

func RegisterSink(sink Sink) {
	if sink == nil {
		return
	}

	globalSink.mutex.Lock()

	globalSink.children = append(globalSink.children, sink)

	globalSink.mutex.Unlock()
}

func Send(event Event) {
	globalSink.Send(event)
}

func WithID(component, id, eventType string) Event {
	return &event{
		EventComponent: component,
		EventData: map[string]interface{}{
			FieldID: id,
		},
		EventType: eventType,
	}
}

//
// Private types
//

// Sink implementation that delegates to child Sinks.
type delegatingSink struct {
	allowedMap map[string]bool
	children   []Sink
	mutex      sync.RWMutex
}

func (sink *delegatingSink) Send(event Event) {
	if event == nil {
		return
	}

	sink.mutex.RLock()
	defer sink.mutex.RUnlock()

	sink.sendInternal(event)
}

func (sink *delegatingSink) sendInternal(event Event) {
	for _, child := range sink.children {
		if globalSink.allowedMap[event.Component()] {
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
// Private variables
//

var (
	globalSink = &delegatingSink{
		allowedMap: make(map[string]bool),
		mutex:      sync.RWMutex{},
	}
)

//
// Private functions
//

func allowFromInternal(component string, shouldAllow bool) {
	globalSink.allowedMap[component] = shouldAllow
}
