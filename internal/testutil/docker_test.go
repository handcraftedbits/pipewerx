package testutil // import "golang.handcraftedbits.com/pipewerx/internal/testutil"

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Testcases
//

// Docker tests

func TestDocker_PhysicalPort(t *testing.T) {
	Convey("When creating a new Docker instance", t, func() {
		var docker = NewDocker("")
		var err error

		defer docker.Destroy()

		err = docker.Run(echoRun)

		So(err, ShouldBeNil)

		Convey("the host port for a non-existent container should not exist", func() {
			So(docker.HostPort("abc", 123), ShouldEqual, -1)
		})

		Convey("the host port for an invalid container port should not exist", func() {
			So(docker.HostPort("echo", 1234), ShouldEqual, -1)
		})
	})
}

func TestDocker_Run(t *testing.T) {
	Convey("When creating a new Docker instance", t, func() {
		var docker = NewDocker("")

		defer docker.Destroy()

		Convey("it should fail to run an invalid Docker image", func() {
			var err = docker.Run(&DockerRun{
				Image: "__invalid__",
				Name:  "invalid",
				Tag:   "latest",
			})

			So(err, ShouldNotBeNil)
		})

		Convey("it should return an error when a timeout occurs waiting for the container port", func() {
			var err error
			var tempDocker = NewDocker("")

			defer tempDocker.Destroy()

			tempDocker.pool.MaxWait = time.Millisecond

			err = tempDocker.Run(&DockerRun{
				Name:  "echo-fail",
				Image: "hashicorp/http-echo",
				Tag:   "latest",
				Args:  []string{"-text", "testing"},
				Port:  1234,
			})

			So(err, ShouldNotBeNil)
		})

		Convey("it should be able to run a simple Docker image", func() {
			var body []byte
			var err = docker.Run(echoRun)
			var reader io.ReadCloser
			var response *http.Response

			So(err, ShouldBeNil)

			response, err = http.Get(fmt.Sprintf("http://localhost:%d", docker.HostPort("echo", 5678)))

			So(err, ShouldBeNil)
			So(response, ShouldNotBeNil)

			reader = response.Body

			So(reader, ShouldNotBeNil)

			body, err = ioutil.ReadAll(reader)

			So(err, ShouldBeNil)
			So(body, ShouldNotBeNil)
			So(strings.TrimSpace(string(body)), ShouldEqual, "testing")

			Convey("and it shouldn't start the container again on a subsequent call to Run()", func() {
				So(docker.resources["echo"], ShouldNotBeNil)
				So(docker.Run(echoRun), ShouldBeNil)
			})
		})
	})
}

func TestNewDocker(t *testing.T) {
	Convey("When creating a new Docker instance", t, func() {
		Convey("it should panic when an invalid endpoint is provided", func() {
			defer func() {
				So(recover(), ShouldNotBeNil)
			}()

			_ = NewDocker("://")
		})
	})
}

// DockerRun tests

func TestDockerRun_AsRunOptions(t *testing.T) {
	Convey("When creating a new DockerRun instance", t, func() {
		var run = &DockerRun{
			Args: []string{"arg1", "arg2"},
			Env: map[string]string{
				"KEY1": "VALUE1",
			},
			Image: "image",
			Tag:   "tag",
			Volumes: map[string]string{
				"/physical1": "/logical1",
			},
		}

		Convey("it should populate all fields correctly", func() {
			var options = run.AsRunOptions()

			So(options, ShouldNotBeNil)
			So(options.Cmd, ShouldHaveLength, 2)
			So(options.Cmd[0], ShouldEqual, "arg1")
			So(options.Cmd[1], ShouldEqual, "arg2")
			So(options.Env, ShouldHaveLength, 1)
			So(options.Env[0], ShouldEqual, "KEY1=VALUE1")
			So(options.Repository, ShouldEqual, "image")
			So(options.Tag, ShouldEqual, "tag")
			So(options.Mounts, ShouldHaveLength, 1)
			So(options.Mounts[0], ShouldEqual, "/physical1:/logical1")
		})
	})
}

//
// Private variables
//

var echoRun = &DockerRun{
	Name:  "echo",
	Image: "hashicorp/http-echo",
	Tag:   "latest",
	Args:  []string{"-text", "testing"},
	Port:  5678,
}
