package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"errors"
)

//
// Public types
//

type Filter interface {
	Source
}

//
// Public functions
//

func NewFilter(name string, newEvaluator NewFileEvaluatorFunc, sources []Source) Filter {
	if newEvaluator == nil {
		newEvaluator = func(Context) (FileEvaluator, error) {
			return &nilFileEvaluator{}, nil
		}
	}

	return &filter{
		input:        newMergedSource(sources),
		name:         name,
		newEvaluator: newEvaluator,
	}
}

//
// Private types
//

type filter struct {
	input        Source
	name         string
	newEvaluator func(Context) (FileEvaluator, error)
}

func (f *filter) Files(context Context) (<-chan Result, func()) {
	var cancel = make(chan struct{})
	var out = make(chan Result)

	go func() {
		var err error
		var evaluator FileEvaluator
		var in <-chan Result
		var sourceCancel func()

		defer func() {
			if evaluator != nil {
				if err := evaluator.Destroy(); err != nil {
					out <- newResult(nil, err)
				}
			}

			close(out)
			close(cancel)
		}()

		if evaluator, err = f.newEvaluator(context); err != nil {
			out <- newResult(nil, err)

			return
		}

		if evaluator == nil {
			out <- newResult(nil, errors.New("nil FileEvaluator was created"))

			return
		}

		in, sourceCancel = f.input.Files(context)

		for result := range in {
			select {
			case <-cancel:
				sourceCancel()

				return

			default:
				var keep bool

				if result.Error() != nil {
					out <- newResult(nil, result.Error())
				}

				keep, err = evaluator.ShouldKeep(result.File())

				if err != nil {
					out <- newResult(nil, err)
				}

				if keep {
					out <- result
				}
			}
		}
	}()

	return out, func() { cancel <- struct{}{} }
}

func (f *filter) Name() string {
	return f.name
}
