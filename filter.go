package pipewerx // import "golang.handcraftedbits.com/pipewerx"

//
// Public types
//

type Filter interface {
	Source
}

type FilterConfig struct {
	Name string
}

//
// Public functions
//

func NewFilter(config FilterConfig, sources []Source, evaluator FileEvaluator) (Filter, error) {
	var err error
	var merged Source

	if evaluator == nil {
		evaluator = &nilFileEvaluator{}
	}

	merged, err = newMergedSource(sources)

	if err != nil {
		return nil, err
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

		defer cancelHelper.finalize()

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

				case <-cancel:
					sourceCancel(nil)

					return
				}
			}
		}
	}()

	return out, cancelHelper.invoker()
}

func (f *filter) Name() string {
	return f.config.Name
}
