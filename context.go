package pipewerx // import "golang.handcraftedbits.com/pipewerx"

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

//
// Public types
//

type Context interface {
	Copy() Context

	Log() *zerolog.Logger

	SetLogLevel(level zerolog.Level)

	Vars() map[string]interface{}
}

type ContextConfig struct {
	Level   zerolog.Level
	UseJSON bool
	Writer  io.Writer
}

//
// Public functions
//

func NewContext(config ContextConfig) Context {
	var logger zerolog.Logger
	var writer io.Writer

	if config.Writer == nil {
		writer = os.Stderr
	} else {
		writer = config.Writer
	}

	if !config.UseJSON {
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
		}
	}

	logger = zerolog.New(writer).With().
		Timestamp().
		Logger().
		Level(config.Level)

	return &context{
		logger: &logger,
		vars:   make(map[string]interface{}),
	}
}

//
// Private types
//

// Context implementation
type context struct {
	logger *zerolog.Logger
	vars   map[string]interface{}
}

func (ctx *context) Copy() Context {
	var newVars = make(map[string]interface{})

	for key, value := range ctx.vars {
		newVars[key] = value
	}

	return &context{
		logger: ctx.logger,
		vars:   newVars,
	}
}

func (ctx *context) Log() *zerolog.Logger {
	return ctx.logger
}

func (ctx *context) SetLogLevel(level zerolog.Level) {
	var logger = ctx.logger.Level(level)

	ctx.logger = &logger
}

func (ctx *context) Vars() map[string]interface{} {
	return ctx.vars
}
