package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/filesystem"
)

//
// Public types
//

type LocalConfig struct {
	Name    string
	Recurse bool
	Root    string
}

//
// Public functions
//

func Local(config LocalConfig) (pipewerx.Source, error) {
	return pipewerx.NewSource(pipewerx.SourceConfig{
		Name:    config.Name,
		Recurse: config.Recurse,
		Root:    config.Root,
	}, filesystem.Local(config.Root))
}
