package testutil // import "golang.handcraftedbits.com/pipewerx/internal/testutil"

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/ory/dockertest/v3"

	. "github.com/smartystreets/goconvey/convey"
)

//
// Public types
//

type Docker struct {
	mutex     sync.Mutex
	pool      *dockertest.Pool
	resources map[string]*dockertest.Resource
}

func (docker *Docker) Destroy() {
	docker.mutex.Lock()
	defer docker.mutex.Unlock()

	for _, resource := range docker.resources {
		_ = docker.pool.Purge(resource)
	}
}

func (docker *Docker) HostPort(name string, containerPort int) int {
	var err error
	var hostPort string
	var resource = docker.resources[name]
	var result int

	if resource == nil {
		return -1
	}

	hostPort = resource.GetPort(fmt.Sprintf("%d/tcp", containerPort))

	result, err = strconv.Atoi(hostPort)

	if err != nil {
		return -1
	}

	return result
}

func (docker *Docker) Run(run *DockerRun) error {
	docker.mutex.Lock()
	defer docker.mutex.Unlock()

	if _, ok := docker.resources[run.Name]; ok {
		return nil
	}

	resource, err := docker.pool.RunWithOptions(run.AsRunOptions())

	if err != nil {
		return err
	}

	docker.resources[run.Name] = resource

	hostPort := docker.HostPort(run.Name, run.Port)

	if run.PingFunc == nil {
		run.PingFunc = func(hostPort int) error {
			connection, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", hostPort))

			if connection != nil {
				_ = connection.Close()
			}

			return err
		}
	}
	err = docker.pool.Retry(func() error {
		return run.PingFunc(hostPort)
	})

	if err != nil {
		return err
	}

	return nil
}

type DockerRun struct {
	Args     []string
	Env      map[string]string
	Image    string
	Name     string
	PingFunc func(int) error
	Port     int
	Tag      string
	Volumes  map[string]string
}

func (run *DockerRun) AsRunOptions() *dockertest.RunOptions {
	var options = &dockertest.RunOptions{
		Cmd:        run.Args,
		Repository: run.Image,
		Tag:        run.Tag,
	}

	if run.Env != nil {
		var i int

		options.Env = make([]string, len(run.Env))

		for key, value := range run.Env {
			options.Env[i] = fmt.Sprintf("%s=%s", key, value)

			i++
		}
	}

	if run.Volumes != nil {
		var i int

		options.Mounts = make([]string, len(run.Volumes))

		for key, value := range run.Volumes {
			options.Mounts[i] = fmt.Sprintf("%s:%s", key, value)

			i++
		}
	}

	return options
}

//
// Public functions
//

func NewDocker(endpoint string) *Docker {
	var docker = &Docker{
		resources: make(map[string]*dockertest.Resource),
	}
	var err error

	docker.pool, err = dockertest.NewPool(endpoint)

	if err != nil {
		panic(err)
	}

	return docker
}

func StartSambaContainer(docker *Docker, absPath string, pingFunc func(int) error) int {
	var err error
	var port int

	err = docker.Run(&DockerRun{
		Args: []string{
			"-s", fmt.Sprintf("%s;/share;yes;yes;no;%s", ConstSMBShare, ConstSMBUser),
			"-u", fmt.Sprintf("%s;%s", ConstSMBUser, ConstSMBPassword),
			"-w", ConstSMBDomain,
		},
		Env:   nil,
		Image: "dperson/samba",
		Name:  "samba",
		Port:  445,
		Tag:   "latest",
		Volumes: map[string]string{
			absPath: "/share",
		},
		PingFunc: pingFunc,
	})

	So(err, ShouldBeNil)

	port = docker.HostPort("samba", 445)

	So(port, ShouldNotEqual, -1)

	return port
}
