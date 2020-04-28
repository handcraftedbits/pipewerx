package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/filesystem"
)

//
// Public types
//

type SMBConfig struct {
	Domain   string
	Host     string
	Name     string
	Password string
	Port     int
	Recurse  bool
	Root     string
	Share    string
	Username string

	enableTestConditions bool
}

//
// Public functions
//

func NewSMB(config SMBConfig) (pipewerx.Source, error) {
	var err error
	var fs pipewerx.Filesystem

	fs, err = filesystem.NewSMB(filesystem.SMBConfig{
		Domain:               config.Domain,
		EnableTestConditions: config.enableTestConditions,
		Host:                 config.Host,
		Password:             config.Password,
		Port:                 config.Port,
		Root:                 config.Root,
		Share:                config.Share,
		Username:             config.Username,
	})

	if err != nil {
		return nil, err
	}

	return pipewerx.NewSource(pipewerx.SourceConfig{
		Name:    config.Name,
		Recurse: config.Recurse,
		Root:    config.Root,
	}, fs)
}
