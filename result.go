package pipewerx // import "golang.handcraftedbits.com/pipewerx"

//
// Public types
//

type Result interface {
	Error() error

	File() File
}

//
// Private types
//

// Result implementation
type result struct {
	err  error
	file File
}

func (res *result) Error() error {
	return res.err
}

func (res *result) File() File {
	return res.file
}

//
// Private functions
//

func newResult(file File, err error) Result {
	return &result{
		err:  err,
		file: file,
	}
}
