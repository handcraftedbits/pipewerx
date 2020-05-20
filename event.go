package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"golang.handcraftedbits.com/pipewerx/internal/event"
)

//
// Private constants
//

const (
	componentFile   = "file"
	componentFilter = "filter"
	componentSource = "source"
)

//
// Private functions
//

func newResultProducedEvent(component, id string, result Result) event.Event {
	var evt = event.WithID(component, id, event.TypeResultProduced)

	if result.Error() != nil {
		evt.Data()[event.FieldError] = result.Error().Error()
	}

	if result.File() != nil {
		evt.Data()[event.FieldFile] = result.File().Path().String()
	}

	return evt
}

// Filter event helpers

func filterEventCancelled(id string) event.Event {
	return event.WithID(componentFilter, id, event.TypeCancelled)
}

func filterEventCreated(id string) event.Event {
	return event.WithID(componentFilter, id, event.TypeCreated)
}

func filterEventDestroyed(id string) event.Event {
	return event.WithID(componentFilter, id, event.TypeDestroyed)
}

func filterEventFinished(id string) event.Event {
	return event.WithID(componentFilter, id, event.TypeFinished)
}

func filterEventResultProduced(id string, result Result) event.Event {
	return newResultProducedEvent(componentFilter, id, result)
}

func filterEventStarted(id string) event.Event {
	return event.WithID(componentFilter, id, event.TypeStarted)
}

// Source event helpers

func sourceEventCancelled(id string) event.Event {
	return event.WithID(componentSource, id, event.TypeCancelled)
}

func sourceEventCreated(id string) event.Event {
	return event.WithID(componentSource, id, event.TypeCreated)
}

func sourceEventDestroyed(id string) event.Event {
	return event.WithID(componentSource, id, event.TypeDestroyed)
}

func sourceEventFinished(id string) event.Event {
	return event.WithID(componentSource, id, event.TypeFinished)
}

func sourceEventResultProduced(id string, result Result) event.Event {
	return newResultProducedEvent(componentSource, id, result)
}

func sourceEventStarted(id string) event.Event {
	return event.WithID(componentSource, id, event.TypeStarted)
}
