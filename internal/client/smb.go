package client // import "golang.handcraftedbits.com/pipewerx/internal/client"

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	pathutil "path"
	"strings"

	"github.com/hirochachacha/go-smb2"
)

//
// Public types
//

type SMB interface {
	Filesystem

	Disconnect()

	OpenFile(path string) (io.ReadCloser, error)
}

type SMBConfig struct {
	Domain   string
	Host     string
	Password string
	Port     int
	Share    string
	Username string
}

//
// Public functions
//

func NewSMB(config *SMBConfig) (SMB, error) {
	var c = &smb{}
	var dialer *smb2.Dialer
	var err error

	c.connection, err = net.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))

	if err != nil {
		return nil, err
	}

	dialer = &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			Domain:   config.Domain,
			User:     config.Username,
			Password: config.Password,
		},
	}

	c.client, err = dialer.Dial(c.connection)

	if err != nil {
		c.Disconnect()

		return nil, err
	}

	c.fs, err = c.client.Mount(fmt.Sprintf("\\\\%s\\%s", config.Host, config.Share))

	if err != nil {
		c.Disconnect()

		return nil, err
	}

	return c, nil
}

//
// Private types
//

// SMB implementation
type smb struct {
	client     *smb2.Client
	connection net.Conn
	fs         *smb2.RemoteFileSystem
}

func (c *smb) AbsolutePath(path string) (string, error) {
	path = strings.ReplaceAll(path, smbPathSeparator, "/")
	path = pathutil.Clean(path)

	return strings.ReplaceAll(path, "/", smbPathSeparator), nil
}

func (c *smb) BasePart(path string) string {
	path = strings.ReplaceAll(path, smbPathSeparator, "/")

	return strings.ReplaceAll(pathutil.Base(path), "/", smbPathSeparator)

}

func (c *smb) DirPart(path string) []string {
	var dir = pathutil.Dir(strings.ReplaceAll(path, smbPathSeparator, "/"))

	if dir == "." {
		// This will be the case for a single file with no directory component, so return an empty array.

		return []string{}
	}

	return strings.Split(dir, "/")
}

// TODO: probably should log as warnings.
func (c *smb) Disconnect() {
	if c.fs != nil {
		_ = c.fs.Umount()
	}

	if c.client != nil {
		_ = c.client.Logoff()
	}

	if c.connection != nil {
		_ = c.connection.Close()
	}
}

func (c *smb) ListFiles(path string) ([]os.FileInfo, error) {
	var err error
	var file *smb2.RemoteFile
	var fileInfos []os.FileInfo

	file, err = c.fs.Open(path)

	if err != nil {
		return nil, err
	}

	fileInfos, err = file.Readdir(-1)

	_ = file.Close()

	if err != nil {
		// TODO: separate client test for this.
		if errors.Is(err, io.EOF) {
			// go-smb2 seems to return this when there's an empty directory.

			return []os.FileInfo{}, nil
		}

		return nil, err
	}

	return fileInfos, nil
}

// TODO: separate client test for this.
func (c *smb) OpenFile(path string) (io.ReadCloser, error) {
	return c.fs.Open(path)
}

func (c *smb) PathSeparator() string {
	return smbPathSeparator
}

func (c *smb) StatFile(path string) (os.FileInfo, error) {
	return c.fs.Stat(path)
}

//
// Private constants
//

const smbPathSeparator = "\\"
