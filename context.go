package pipewerx // import "golang.handcraftedbits.com/pipewerx"

//
// Public types
//

type Context interface {
	Copy() Context

	Vars() map[string]interface{}
}

//
// Private types
//

// Context implementation
type context struct {
	vars map[string]interface{}
}

func (ctx *context) Copy() Context {
	var newVars = make(map[string]interface{})

	for key, value := range ctx.vars {
		newVars[key] = value
	}

	return &context{
		vars: newVars,
	}
}

func (ctx *context) Vars() map[string]interface{} {
	return ctx.vars
}

//
// Private functions
//

func newContext() Context {
	return &context{
		vars: make(map[string]interface{}),
	}
}
