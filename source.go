package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"sync"
)

//
// Public types
//

type Source interface {
	Files(context Context) (<-chan Result, CancelFunc)

	ID() string

	destroy() error
}

type SourceConfig struct {
	ID      string
	Recurse bool
	Root    string
}

//
// Public functions
//

func NewSource(config SourceConfig, fs Filesystem) (Source, error) {
	var err error

	if fs == nil {
		return nil, errSourceNilFilesystem
	}

	err = validateID(config.ID)

	if err != nil {
		return nil, err
	}

	if eventAllowedFrom(componentSource) {
		sendEvent(sourceEventCreated(config.ID))
	}

	return &source{
		config: config,
		fs:     fs,
	}, nil
}

//
// Private types
//

// Source implementation that combines multiple Sources
type mergedSource struct {
	id      string
	sources []Source
}

func (merged *mergedSource) Files(context Context) (<-chan Result, CancelFunc) {
	var cancel = make(chan struct{})
	var cancelHelper *cancellationHelper
	var out = make(chan Result)
	var wg sync.WaitGroup

	cancelHelper = newCancellationHelper(context.Log(), out, cancel, &wg)

	wg.Add(len(merged.sources))

	go cancelHelper.finalize()

	for i, source := range merged.sources {
		go func(index int, self Source) {
			var in <-chan Result
			var sourceCancel CancelFunc

			defer func() {
				wg.Done()
			}()

			in, sourceCancel = self.Files(context)

			for item := range in {
				select {
				case out <- item:

				case <-cancel:
					sourceCancel(nil)

					return
				}
			}
		}(i, source)
	}

	return out, cancelHelper.invoker()
}

func (merged *mergedSource) ID() string {
	return merged.id
}

func (merged *mergedSource) destroy() error {
	var errs []error

	for _, source := range merged.sources {
		if err := source.destroy(); err != nil {
			errs = append(errs, err)
		}
	}

	if errs != nil {
		return newMultiError("an error occurred while destroying the source", errs)
	}

	return nil
}

// Default Source implementation
type source struct {
	config SourceConfig
	fs     Filesystem
}

func (src *source) Files(context Context) (<-chan Result, CancelFunc) {
	var cancel = make(chan struct{})
	var cancelHelper *cancellationHelper
	var out = make(chan Result)

	cancelHelper = newCancellationHelper(context.Log(), out, cancel, nil)

	go func() {
		var err error
		var file File
		var res *result
		var stepper *pathStepper

		if eventAllowedFrom(componentSource) {
			sendEvent(sourceEventStarted(src.config.ID))
		}

		defer func() {
			if eventAllowedFrom(componentSource) {
				sendEvent(sourceEventFinished(src.config.ID))
			}

			cancelHelper.finalize()
		}()

		if stepper, err = newPathStepper(src.fs, src.config.Root, src.config.Recurse); err != nil {
			res = &result{
				err: err,
			}

			out <- res

			if eventAllowedFrom(componentSource) {
				sendEvent(sourceEventResultProduced(src.config.ID, res))
			}

			return
		}

		for {
			file, err = stepper.nextFile()

			if file == nil && err == nil {
				return
			}

			res = &result{
				err:  err,
				file: file,
			}

			select {
			case out <- res:
				if eventAllowedFrom(componentSource) {
					sendEvent(sourceEventResultProduced(src.config.ID, res))
				}

			case <-cancel:
				if eventAllowedFrom(componentSource) {
					sendEvent(sourceEventCancelled(src.config.ID))
				}

				return
			}
		}
	}()

	return out, cancelHelper.invoker()
}

func (src *source) ID() string {
	return src.config.ID
}

func (src *source) destroy() error {
	if eventAllowedFrom(componentSource) {
		sendEvent(sourceEventDestroyed(src.config.ID))
	}

	return src.fs.Destroy()
}

//
// Private constants
//

const componentSource = "source"

//
// Private functions
//

func newMergedSource(id string, sources []Source) (Source, error) {
	var duplicates = make(map[Source]bool)
	var sanitizedSources = make([]Source, 0)

	for _, src := range sources {
		// Guard against Source instances being used more than once in the same merged Source -- that will likely lead
		// to issues.

		if _, ok := duplicates[src]; ok {
			continue
		} else {
			duplicates[src] = true
		}

		// We also need to guard against nil Sources being used.

		if src != nil {
			sanitizedSources = append(sanitizedSources, src)
		}
	}

	switch len(sanitizedSources) {
	case 0:
		// No valid sources provided, so return an error.

		return nil, errSourceNone

	case 1:
		// Only a single source provided, so just use it directly.

		return sanitizedSources[0], nil
	}

	return &mergedSource{
		id:      id,
		sources: sanitizedSources,
	}, nil
}

func sourceEventCancelled(id string) Event {
	return newEvent(componentSource, id, eventTypeCancelled)
}

func sourceEventCreated(id string) Event {
	return newEvent(componentSource, id, eventTypeCreated)
}

func sourceEventDestroyed(id string) Event {
	return newEvent(componentSource, id, eventTypeDestroyed)
}

func sourceEventFinished(id string) Event {
	return newEvent(componentSource, id, eventTypeFinished)
}

func sourceEventResultProduced(id string, result Result) Event {
	return newResultProducedEvent(componentSource, id, result)
}

func sourceEventStarted(id string) Event {
	return newEvent(componentSource, id, eventTypeStarted)
}
