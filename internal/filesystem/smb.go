package filesystem // import "golang.handcraftedbits.com/pipewerx/internal/filesystem"

/*
#cgo pkg-config: smbclient

#include "smb_native.h"
*/
import "C"

import (
	"fmt"
	"io"
	"os"
	pathutil "path"
	"time"
	"unsafe"

	"golang.handcraftedbits.com/pipewerx"
)

//
// Public types
//

type SMBConfig struct {
	Domain               string
	EnableTestConditions bool
	Host                 string
	Password             string
	Port                 int
	Root                 string
	Share                string
	Username             string
}

//
// Public functions
//

func SMB(config SMBConfig) (pipewerx.Filesystem, error) {
	var cContext *C.SMBCCTX
	var cDomain = C.CString(config.Domain)
	var cPassword = C.CString(config.Password)
	var cUsername = C.CString(config.Username)
	var err error

	cContext, err = C.pipewerx_smb_create_context(cDomain, cUsername, cPassword, C.bool(config.EnableTestConditions))

	if cContext == nil {
		C.free(unsafe.Pointer(cDomain))
		C.free(unsafe.Pointer(cPassword))
		C.free(unsafe.Pointer(cUsername))

		return nil, err
	}

	return &smb{
		config:   config,
		cContext: cContext,
	}, nil
}

//
// Private types
//

// SMB pipewerx.Filesystem implementation
type smb struct {
	pipewerx.FilesystemDefaults

	config   SMBConfig
	cContext *C.SMBCCTX
}

func (fs *smb) Destroy() error {
	var err error
	var cRet C.int

	cRet, err = C.pipewerx_smb_destroy_context(fs.cContext, C.bool(fs.config.EnableTestConditions))

	if int(cRet) != 0 {
		return err
	}

	return nil
}

func (fs *smb) ListFiles(path string) ([]os.FileInfo, error) {
	var cDirHandle *C.SMBCFILE
	var cURL = C.CString(fs.makeURL(path, false))
	var err error
	var fileInfos = make([]os.FileInfo, 0)

	defer C.free(unsafe.Pointer(cURL))

	cDirHandle, err = C.pipewerx_smb_opendir(fs.cContext, cURL)

	if cDirHandle == nil {
		return nil, err
	}

	defer C.pipewerx_smb_closedir(fs.cContext, cDirHandle)

	for {
		var cFileInfo *C.struct_libsmb_file_info
		var cStat C.struct_stat
		var name string

		cFileInfo, err = C.pipewerx_smb_readdirplus2(fs.cContext, cDirHandle, &cStat,
			C.bool(fs.config.EnableTestConditions))

		if cFileInfo == nil {
			if err != nil {
				return nil, err
			}

			// No error, so this is simply the end of the listing.

			return fileInfos, nil
		}

		name = C.GoString(cFileInfo.name)

		// libsmbclient returns "." and "..", which we don't want.  Filter them out.

		if name != "." && name != ".." {
			fileInfos = append(fileInfos, newSMBFileInfo(C.GoString(cFileInfo.name), &cStat))
		}
	}
}

func (fs *smb) ReadFile(path string) (io.ReadCloser, error) {
	var cFileHandle *C.SMBCFILE
	var cURL = C.CString(fs.makeURL(path, true))
	var err error

	defer C.free(unsafe.Pointer(cURL))

	cFileHandle, err = C.pipewerx_smb_open(fs.cContext, cURL, C.int(os.O_RDONLY), C.mode_t(0))

	if cFileHandle == nil {
		return nil, err
	}

	return &smbReadCloser{
		cContext:    fs.cContext,
		cFileHandle: cFileHandle,
	}, nil
}

func (fs *smb) StatFile(path string) (os.FileInfo, error) {
	var cRet C.int
	var cStat C.struct_stat
	var cURL = C.CString(fs.makeURL(path, false))
	var err error

	defer C.free(unsafe.Pointer(cURL))

	cRet, err = C.pipewerx_smb_stat(fs.cContext, cURL, &cStat)

	if int(cRet) != 0 {
		return nil, err
	}

	return newSMBFileInfo(path, &cStat), nil
}

func (fs *smb) makeURL(path string, includeRoot bool) string {
	if includeRoot && (fs.config.Root != "" && fs.config.Root != path) {
		path = fs.config.Root + "/" + path
	}

	return fmt.Sprintf("smb://%s:%d/%s/%s", fs.config.Host, fs.config.Port, fs.config.Share, pathutil.Clean(path))
}

// SMB io.ReadCloser implementation
type smbReadCloser struct {
	cContext    *C.SMBCCTX
	cFileHandle *C.SMBCFILE
}

func (reader *smbReadCloser) Close() error {
	var cRet C.int
	var err error

	cRet, err = C.pipewerx_smb_close(reader.cContext, reader.cFileHandle)

	if int(cRet) != 0 {
		return err
	}

	return nil
}

func (reader *smbReadCloser) Read(p []byte) (int, error) {
	var bytesRead int
	var err error
	var read C.ssize_t

	if len(p) == 0 {
		return 0, nil
	}

	read, err = C.pipewerx_smb_read(reader.cContext, reader.cFileHandle, unsafe.Pointer(&p[0]), C.size_t(len(p)))

	bytesRead = int(read)

	if bytesRead < 0 {
		return bytesRead, err
	} else if bytesRead > 0 {
		return bytesRead, nil
	}

	return bytesRead, io.EOF
}

//
// Private functions
//

func newSMBFileInfo(path string, cStat *C.struct_stat) os.FileInfo {
	var mode = os.FileMode(cStat.st_mode)

	// os.FileMode doesn't use the same directory mask as stat does, so if we find the stat directory mask
	// (S_IFDIR = 040000), we have to change the mode to use the os.FileMode directory mask.

	if mode&040000 != 0 {
		mode |= os.ModeDir
	}

	return &fileInfo{
		mode:    mode & (os.ModeDir | os.ModePerm),
		modTime: time.Unix(int64(cStat.st_mtim.tv_sec), int64(cStat.st_mtim.tv_nsec)),
		name:    path,
		size:    int64(cStat.st_size),
	}
}
