package client // import "golang.handcraftedbits.com/pipewerx/internal/client"

import (
	"os"
	"testing"

	"golang.handcraftedbits.com/pipewerx/internal/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// SMB tests

func TestSMB_AbsolutePath(t *testing.T) {
	Convey("A SMB client should return proper absolute paths", t, func() {
		var client = &smb{}
		var err error
		var result string

		result, err = client.AbsolutePath("abs")
		So(err, ShouldBeNil)
		So(result, ShouldEqual, "abs")

		result, err = client.AbsolutePath("a\\b\\c")
		So(err, ShouldBeNil)
		So(result, ShouldEqual, "a\\b\\c")

		result, err = client.AbsolutePath("a\\b\\..\\c")
		So(err, ShouldBeNil)
		So(result, ShouldEqual, "a\\c")
	})
}

func TestSMB_BasePart(t *testing.T) {
	Convey("A SMB client should return the proper base part for a file path", t, func() {
		var client = &smb{}

		So(client.BasePart("abc"), ShouldEqual, "abc")
		So(client.BasePart("a\\b\\c\\abc"), ShouldEqual, "abc")
	})
}

func TestSMB_ConnectAndMount(t *testing.T) {
	// TODO: need invalid username, password, domain, share tests

	Convey("A SMB client should be able to connect and disconnect successfully", t, func() {
		var client SMB
		var port = startSambaContainer()

		client = newSMBClient(port)

		client.Disconnect()
	})
}

func TestSMB_DirPart(t *testing.T) {
	Convey("A SMB client should return the proper directory part for a file path", t, func() {
		var client = &smb{}
		var result []string

		result = client.DirPart("abc")

		So(result, ShouldNotBeNil)
		So(result, ShouldHaveLength, 0)

		result = client.DirPart("a\\b\\c\\abc")

		So(result, ShouldNotBeNil)
		So(result, ShouldHaveLength, 3)
		So(result[0], ShouldEqual, "a")
		So(result[1], ShouldEqual, "b")
		So(result[2], ShouldEqual, "c")
	})
}

func TestSMB_ListFiles(t *testing.T) {
	Convey("When creating a SMB client", t, func() {
		var client SMB
		var err error
		var port = startSambaContainer()

		client = newSMBClient(port)

		defer client.Disconnect()

		Convey("it should return an error when an invalid directory is specified", func() {
			var fileInfos []os.FileInfo

			fileInfos, err = client.ListFiles(";;;;")

			So(fileInfos, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("it should be able to list files in a valid directory", func() {
			var fileInfo os.FileInfo
			var fileInfos []os.FileInfo
			var fileInfosByMap = make(map[string]os.FileInfo)

			fileInfos, err = client.ListFiles("")

			So(err, ShouldBeNil)
			So(fileInfos, ShouldNotBeNil)
			So(fileInfos, ShouldHaveLength, 3)

			for _, fileInfo = range fileInfos {
				fileInfosByMap[fileInfo.Name()] = fileInfo
			}

			fileInfo = fileInfosByMap["a.test"]

			So(fileInfo, ShouldNotBeNil)
			So(fileInfo.IsDir(), ShouldBeFalse)

			fileInfo = fileInfosByMap["b.test"]

			So(fileInfo, ShouldNotBeNil)
			So(fileInfo.IsDir(), ShouldBeFalse)

			fileInfo = fileInfosByMap["c"]

			So(fileInfo, ShouldNotBeNil)
			So(fileInfo.IsDir(), ShouldBeTrue)
		})
	})
}

func TestSMB_PathSeparator(t *testing.T) {
	Convey("A SMB client should have the correct path separator", t, func() {
		var client = &smb{}

		So(client.PathSeparator(), ShouldEqual, smbPathSeparator)
	})
}

func TestSMB_StatFile(t *testing.T) {
	Convey("When creating a SMB client", t, func() {
		var client SMB
		var err error
		var port = startSambaContainer()

		client = newSMBClient(port)

		defer client.Disconnect()

		Convey("it should return an error when an invalid file is specified", func() {
			var fileInfo os.FileInfo

			fileInfo, err = client.StatFile(";;;;")

			So(fileInfo, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("it should be able to stat a valid file", func() {
			var fileInfo os.FileInfo

			fileInfo, err = client.StatFile("a.test")

			So(err, ShouldBeNil)
			So(fileInfo, ShouldNotBeNil)
			So(fileInfo.Name(), ShouldEqual, "a.test")
			So(fileInfo.IsDir(), ShouldBeFalse)
		})
	})
}

func TestNewSMB(t *testing.T) {
	Convey("When creating a SMB client", t, func() {
		Convey("it should return an error when an invalid host or port is provided", func() {
			var client SMB
			var config = &SMBConfig{
				Host: "????",
			}
			var err error

			config.Host = "localhost"
			config.Port = -1

			client, err = NewSMB(config)

			So(client, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("it should return an error when an invalid share name is provided", func() {
			var client SMB
			var config *SMBConfig
			var err error
			var port = startSambaContainer()

			config = newSMBConfig(port)

			config.Share = "xxxx"

			client, err = NewSMB(config)

			So(client, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}

//
// Private functions
//

func newSMBClient(port int) SMB {
	client, err := NewSMB(newSMBConfig(port))

	So(client, ShouldNotBeNil)
	So(err, ShouldBeNil)

	return client
}

func newSMBConfig(port int) *SMBConfig {
	return &SMBConfig{
		Domain:   testutil.ConstSMBDomain,
		Host:     "localhost",
		Password: testutil.ConstSMBPassword,
		Port:     port,
		Share:    testutil.ConstSMBShare,
		Username: testutil.ConstSMBUser,
	}
}

func startSambaContainer() int {
	return testutil.StartSambaContainer(docker, "testdata/mountDir", func(hostPort int) error {
		client, clientError := NewSMB(newSMBConfig(hostPort))

		if client != nil {
			client.Disconnect()
		}

		return clientError
	})
}
