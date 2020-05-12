package pipewerx // import "golang.handcraftedbits.com/pipewerx"

//
// Public types
//

type Filter interface {
	Source
}

type FilterConfig struct {
	ID string
}

//
// Public functions
//

func NewFilter(config FilterConfig, sources []Source, evaluator FileEvaluator) (Filter, error) {
	var err error
	var merged Source

	err = validateID(config.ID)

	if err != nil {
		return nil, err
	}

	if evaluator == nil {
		evaluator = &nilFileEvaluator{}
	}

	merged, err = newMergedSource(config.ID, sources)

	if err != nil {
		return nil, err
	}

	if eventAllowedFrom(componentFilter) {
		sendEvent(filterEventCreated(config.ID))
	}

	return &filter{
		config:    config,
		evaluator: evaluator,
		input:     merged,
	}, nil
}

//
// Private types
//

// Filter implementation
type filter struct {
	config    FilterConfig
	evaluator FileEvaluator
	input     Source
}

func (f *filter) destroy() error {
	if eventAllowedFrom(componentFilter) {
		sendEvent(filterEventDestroyed(f.ID()))
	}

	return f.evaluator.Destroy()
}

func (f *filter) Files(context Context) (<-chan Result, CancelFunc) {
	var cancel = make(chan struct{})
	var cancelHelper *cancellationHelper
	var out = make(chan Result)

	cancelHelper = newCancellationHelper(context.Log(), out, cancel, nil)

	go func() {
		var err error
		var in <-chan Result
		var sourceCancel CancelFunc

		if eventAllowedFrom(componentFilter) {
			sendEvent(filterEventStarted(f.ID()))
		}

		defer func() {
			if eventAllowedFrom(componentFilter) {
				sendEvent(filterEventFinished(f.ID()))
			}

			cancelHelper.finalize()
		}()

		in, sourceCancel = f.input.Files(context)

		for res := range in {
			var keep bool

			if res.Error() == nil {
				func() {
					defer func() {
						if value := recover(); value != nil {
							err = newPanicError(value)
							keep = false
						}
					}()

					keep, err = f.evaluator.ShouldKeep(res.File())
				}()

				if err != nil {
					res = &result{err: err, file: nil}
				}
			}

			if keep || res.Error() != nil {
				select {
				case out <- res:
					if eventAllowedFrom(componentFilter) {
						sendEvent(filterEventResultProduced(f.ID(), res))
					}

				case <-cancel:
					if eventAllowedFrom(componentFilter) {
						sendEvent(filterEventCancelled(f.ID()))
					}

					sourceCancel(nil)

					return
				}
			}
		}
	}()

	return out, cancelHelper.invoker()
}

func (f *filter) ID() string {
	return f.config.ID
}

//
// Private constants
//

const componentFilter = "filter"

//
// Private functions
//

func filterEventCancelled(id string) Event {
	return newEvent(componentFilter, id, eventTypeCancelled)
}

func filterEventCreated(id string) Event {
	return newEvent(componentFilter, id, eventTypeCreated)
}

func filterEventDestroyed(id string) Event {
	return newEvent(componentFilter, id, eventTypeDestroyed)
}

func filterEventFinished(id string) Event {
	return newEvent(componentFilter, id, eventTypeFinished)
}

func filterEventResultProduced(id string, result Result) Event {
	return newResultProducedEvent(componentFilter, id, result)
}

func filterEventStarted(id string) Event {
	return newEvent(componentFilter, id, eventTypeStarted)
}
