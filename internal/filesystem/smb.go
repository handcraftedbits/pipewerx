package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	pathutil "path"
	"strings"

	"github.com/hirochachacha/go-smb2"

	"golang.handcraftedbits.com/pipewerx"
)

//
// Public types
//

type SMBConfig struct {
	Domain   string
	Host     string
	Password string
	Port     int
	Root     string
	Share    string
	Username string
}

//
// Public functions
//

func NewSMB(config SMBConfig) (pipewerx.Filesystem, error) {
	var dialer *smb2.Dialer
	var err error
	var fs = &smb{
		root: config.Root,
	}

	fs.connection, err = net.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))

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

	fs.client, err = dialer.Dial(fs.connection)

	if err != nil {
		// TODO: log.
		_ = fs.Destroy()

		return nil, err
	}

	fs.remote, err = fs.client.Mount(fmt.Sprintf("\\\\%s\\%s", config.Host, config.Share))

	if err != nil {
		// TODO: log
		_ = fs.Destroy()

		return nil, err
	}

	return fs, nil
}

//
// Private types
//

// pipewerx.Filesystem implementation for an SMB filesystem
type smb struct {
	client     *smb2.Client
	connection net.Conn
	remote     *smb2.RemoteFileSystem
	root       string
}

func (fs *smb) AbsolutePath(path string) (string, error) {
	path = strings.ReplaceAll(path, smbPathSeparator, "/")
	path = pathutil.Clean(path)

	return strings.ReplaceAll(path, "/", smbPathSeparator), nil
}

func (fs *smb) BasePart(path string) string {
	path = strings.ReplaceAll(path, smbPathSeparator, "/")

	return strings.ReplaceAll(pathutil.Base(path), "/", smbPathSeparator)
}

func (fs *smb) Destroy() error {
	if fs.remote != nil {
		_ = fs.remote.Umount()
	}

	if fs.client != nil {
		_ = fs.client.Logoff()
	}

	if fs.connection != nil {
		_ = fs.connection.Close()
	}

	// TODO: proper error.

	return nil
}

func (fs *smb) DirPart(path string) []string {
	var dir = pathutil.Dir(strings.ReplaceAll(path, smbPathSeparator, "/"))

	if dir == "." {
		// This will be the case for a single file with no directory component, so return an empty array.

		return []string{}
	}

	return strings.Split(dir, "/")
}

func (fs *smb) ListFiles(path string) ([]os.FileInfo, error) {
	var err error
	var file *smb2.RemoteFile
	var fileInfos []os.FileInfo

	file, err = fs.remote.Open(path)

	if err != nil {
		return nil, err
	}

	fileInfos, err = file.Readdir(-1)

	_ = file.Close()

	if err != nil {
		if errors.Is(err, io.EOF) {
			// go-smb2 seems to return this when there's an empty directory.

			return []os.FileInfo{}, nil
		}

		return nil, err
	}

	return fileInfos, nil
}

func (fs *smb) PathSeparator() string {
	return smbPathSeparator
}

func (fs *smb) ReadFile(path string) (io.ReadCloser, error) {
	// TODO: ugly, gotta be a better way.
	if fs.root != "" {
		if fs.BasePart(fs.root) != path {
			path = fs.root + smbPathSeparator + path
		} else {
			path = fs.root
		}
	}

	return fs.remote.Open(path)
}

func (fs *smb) StatFile(path string) (os.FileInfo, error) {
	return fs.remote.Stat(path)
}

//
// Private constants
//

const smbPathSeparator = "\\"
