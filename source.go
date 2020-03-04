package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"
	"sync"
)

//
// Public types
//

type Source interface {
	Files(context Context) (<-chan Result, func())

	Name() string
}

//
// Public functions
//

func NewSource(name string, newProducer func(Context) (FileProducer, error)) Source {
	if newProducer == nil {
		newProducer = func(context Context) (FileProducer, error) {
			return &nilFileProducer{}, nil
		}
	}

	return &source{
		name:        name,
		newProducer: newProducer,
	}
}

//
// Private types
//

// Source implementation that combines multiple Sources
type mergedSource struct {
	name    string
	sources []Source
}

func (merged *mergedSource) Files(context Context) (<-chan Result, func()) {
	var cancel = make(chan struct{})
	var out = make(chan Result)
	var wg sync.WaitGroup

	wg.Add(len(merged.sources))

	for i, source := range merged.sources {
		go func(index int, self Source) {
			var in <-chan Result
			var sourceCancel func()

			defer func() {
				wg.Done()
			}()

			in, sourceCancel = self.Files(context)

			for item := range in {
				select {
				case out <- item:

				case <-cancel:
					sourceCancel()

					return
				}
			}
		}(i, source)
	}

	go func() {
		wg.Wait()

		close(out)
	}()

	return out, func() { close(cancel) }
}

func (merged *mergedSource) Name() string {
	return merged.name
}

// Default Source implementation
type source struct {
	name        string
	newProducer func(Context) (FileProducer, error)
}

func (src *source) Files(context Context) (<-chan Result, func()) {
	var cancel = make(chan struct{})
	var out = make(chan Result)

	go func() {
		var err error
		var producer FileProducer

		defer func() {
			if producer != nil {
				if err := producer.Destroy(); err != nil {
					out <- NewResult(nil, err)
				}
			}

			close(out)
		}()

		if producer, err = src.newProducer(context); err != nil {
			out <- NewResult(nil, err)

			return
		}

		if producer == nil {
			// TODO: proper message
			out <- NewResult(nil, errors.New("nil Producer"))

			return
		}

		for {
			select {
			case out <- NewResult(producer.Next()):

			case <-cancel:
				return
			}
		}
	}()

	return out, func() { close(cancel) }
}

func (src *source) Name() string {
	return src.name
}

//
// Private functions
//

func newMergedSource(sources ...Source) Source {
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
		// No valid sources provided, so just create an empty one.

		// TODO: nls
		return NewSource("<empty source>", nil)

	case 1:
		// Only a single source provided, so just use it directly.

		return sanitizedSources[0]
	}

	return &mergedSource{
		// TODO: nls
		name:    "<merged source>",
		sources: sanitizedSources,
	}
}
