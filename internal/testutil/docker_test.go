package testutil // import "golang.handcraftedbits.com/pipewerx/internal/testutil"

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//
// Testcases
//

// Docker tests

var _ = Describe("Docker", func() {
	Describe("given a new instance", func() {
		Describe("calling grepLogs", func() {
			var err error
			var result bool

			Context("with a non-existent resource", func() {
				It("should return an error", func() {
					result, err = docker.grepLogs("abc", "")

					Expect(result).To(BeFalse())
					Expect(err).NotTo(BeNil())
				})
			})

			Context("with an invalid regular expression", func() {
				It("should return an error", func() {
					err = docker.Run(echoRun)

					Expect(err).To(BeNil())

					result, err = docker.grepLogs("echo", "*")

					Expect(result).To(BeFalse())
					Expect(err).NotTo(BeNil())
				})
			})

			Context("with a non-existent container", func() {
				var containerID string

				JustBeforeEach(func() {
					containerID = ""
				})

				JustAfterEach(func() {
					if containerID != "" {
						docker.resources["echo"].Container.ID = containerID
					}
				})

				It("should return an error", func() {
					err = docker.Run(echoRun)

					Expect(err).To(BeNil())

					containerID = docker.resources["echo"].Container.ID

					docker.resources["dummy"] = docker.resources["echo"]
					docker.resources["dummy"].Container.ID = "xyz"

					result, err = docker.grepLogs("echo", "abc")

					Expect(result).To(BeFalse())
					Expect(err).NotTo(BeNil())
				})
			})
		})

		Describe("calling HostPort", func() {
			Context("with a non-existent resource", func() {
				It("should return -1", func() {
					Expect(docker.HostPort("abc", 123)).To(Equal(-1))
				})
			})

			Context("with an invalid container port", func() {
				JustBeforeEach(func() {
					var err = docker.Run(echoRun)

					Expect(err).To(BeNil())
				})

				It("should return -1", func() {
					Expect(docker.HostPort("echo", 1234)).To(Equal(-1))
				})
			})
		})

		Describe("calling Run", func() {
			var err error

			Context("with an invalid Docker image", func() {
				It("should return an error", func() {
					err = docker.Run(&DockerRun{
						Image: "__invalid__",
						Name:  "invalid",
						Tag:   "latest",
					})

					Expect(err).NotTo(BeNil())
				})
			})

			Context("when a timeout occurs waiting for the container port", func() {
				It("should return an error", func() {
					docker.pool.MaxWait = time.Millisecond

					err = docker.Run(&DockerRun{
						Name:  "echo-fail",
						Image: "hashicorp/http-echo",
						Tag:   "latest",
						Args:  []string{"-text", "testing"},
						Port:  1234,
					})

					Expect(err).NotTo(BeNil())
				})
			})

			Context("with a valid Docker image", func() {
				It("should succeed and shouldn't start the container again on a subsequent call", func() {
					var body []byte
					var reader io.ReadCloser
					var response *http.Response

					err = docker.Run(echoRun)

					Expect(err).To(BeNil())

					response, err = http.Get(fmt.Sprintf("http://localhost:%d", docker.HostPort("echo", 5678)))

					Expect(err).To(BeNil())
					Expect(response).NotTo(BeNil())

					reader = response.Body

					Expect(reader).NotTo(BeNil())

					body, err = ioutil.ReadAll(reader)

					Expect(err).To(BeNil())
					Expect(body).NotTo(BeNil())
					Expect(strings.TrimSpace(string(body))).To(Equal("testing"))

					Expect(docker.resources["echo"]).NotTo(BeNil())
					Expect(docker.Run(echoRun)).To(BeNil())
				})
			})
		})
	})
})

var _ = Describe("NewDocker", func() {
	Describe("calling NewDocker", func() {
		Context("with an invalid endpoint", func() {
			It("should panic", func() {
				defer func() {
					Expect(recover()).NotTo(BeNil())
				}()

				_ = NewDocker("://")
			})
		})
	})
})

// DockerRun tests

var _ = Describe("DockerRun", func() {
	Describe("given a new instance", func() {
		var run *DockerRun

		BeforeEach(func() {
			run = &DockerRun{
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
		})

		Describe("calling AsRunOptions", func() {
			It("should return a dockertest.RunOptions object that has all fields correctly populated", func() {
				var options = run.AsRunOptions()

				Expect(options).NotTo(BeNil())
				Expect(options.Cmd).NotTo(BeNil())
				Expect(options.Cmd).To(HaveLen(2))
				Expect(options.Cmd[0]).To(Equal("arg1"))
				Expect(options.Cmd[1]).To(Equal("arg2"))
				Expect(options.Env).NotTo(BeNil())
				Expect(options.Env).To(HaveLen(1))
				Expect(options.Env[0]).To(Equal("KEY1=VALUE1"))
				Expect(options.Mounts).NotTo(BeNil())
				Expect(options.Mounts).To(HaveLen(1))
				Expect(options.Mounts[0]).To(Equal("/physical1:/logical1"))
				Expect(options.Repository).To(Equal("image"))
				Expect(options.Tag).To(Equal("tag"))
			})
		})
	})
})

// Tests for container helpers

var _ = Describe("StartSambaContainer", func() {
	Describe("calling StartSambaContainer", func() {
		It("should succeed", func() {
			StartSambaContainer(docker, TestdataPathFilesystem)
		})
	})
})

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
