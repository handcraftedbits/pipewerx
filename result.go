package pipewerx // import "golang.handcraftedbits.com/pipewerx"

//
// Public types
//

type Result interface {
	Error() error

	File() File
}

//
// Public functions
//

func NewResult(file File, err error) Result {
	return &result{
		err:  err,
		file: file,
	}
}

//
// Private types
//

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
