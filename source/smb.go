package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"io"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/client"
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
}

//
// Public functions
//

func NewSMB(config *SMBConfig) pipewerx.Source {
	return pipewerx.NewSource(config.Name, func(pipewerx.Context) (pipewerx.FileProducer, error) {
		var smbClient client.SMB
		var clientConfig *client.SMBConfig
		var err error
		var stepper *pathStepper

		clientConfig = &client.SMBConfig{
			Domain:   config.Domain,
			Host:     config.Host,
			Password: config.Password,
			Port:     config.Port,
			Share:    config.Share,
			Username: config.Username,
		}

		smbClient, err = client.NewSMB(clientConfig)

		if err != nil {
			return nil, err
		}

		stepper, err = newPathStepper(smbClient, config.Root, config.Recurse)

		if err != nil {
			return nil, err
		}

		return &smbFileProducer{
			smbClient: smbClient,
			stepper:   stepper,
		}, nil
	})
}

//
// Private types
//

// SMB pipewerx.File implementation
type smbFile struct {
	smbClient client.SMB
	path      pipewerx.FilePath
	realPath  string
}

func (file *smbFile) Path() pipewerx.FilePath {
	return file.path
}

func (file *smbFile) Reader() (io.ReadCloser, error) {
	return file.smbClient.OpenFile(file.realPath)
}

// SMB pipewerx.FileProducer implementation
type smbFileProducer struct {
	smbClient client.SMB
	stepper   *pathStepper
}

// TODO: revisit.  Probably don't need/want destructors on FileProducers at all?  Should have it at the Source level?
func (producer *smbFileProducer) Destroy() error {
	//producer.smbClient.Disconnect()

	return nil
}

func (producer *smbFileProducer) Next() (pipewerx.File, error) {
	var err error
	var path pipewerx.FilePath
	var root string

	path, err = producer.stepper.nextFile()

	if err != nil {
		return nil, err
	}

	if path == nil {
		return nil, nil
	}

	root = producer.stepper.root

	// go-smb2 doesn't like paths that start with the path separator, so if we're in the share root (i.e., path is
	// empty), we don't have to tack on a path separator.

	if root != "" {
		root += producer.stepper.fs.PathSeparator()
	}

	return &smbFile{
		smbClient: producer.smbClient,
		path:      path,
		realPath:  root + path.String(),
	}, nil
}
