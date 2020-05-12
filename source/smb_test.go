package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"golang.handcraftedbits.com/pipewerx"
	"golang.handcraftedbits.com/pipewerx/internal/testutil"
)

//
// Testcases
//

// SMB Source tests

var _ = Describe("SMB Source", func() {
	Describe("calling SMB", func() {
		Context("with test conditions enabled", func() {
			It("should return an error", func() {
				var err error
				var source pipewerx.Source

				source, err = SMB(SMBConfig{
					enableTestConditions: true,
				})

				Expect(source).To(BeNil())
				Expect(err).NotTo(BeNil())
			})
		})
	})
})

var _ = testSource(testSourceConfig{
	createFunc: func(id, root string, recurse bool) (pipewerx.Source, error) {
		var config = newSMBConfig(portSamba)

		config.ID = id
		config.Recurse = recurse
		config.Root = root

		return SMB(config)
	},
	name:          "SMB",
	pathSeparator: "/",
	realPath: func(root, path string) string {
		return path
	},
})

//
// Private functions
//

func newSMBConfig(port int) SMBConfig {
	return SMBConfig{
		Domain:   testutil.ConstSMBDomain,
		Host:     "localhost",
		Password: testutil.ConstSMBPassword,
		Port:     port,
		Share:    testutil.ConstSMBShare,
		Username: testutil.ConstSMBUser,
	}
}
