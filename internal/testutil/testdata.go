package testutil // import "golang.handcraftedbits.com/pipewerx/internal/testutil"

import (
	"os"
	"path/filepath"
	"strings"
)

//
// Public variables
//

var (
	TestdataPathFilesystem string
)

//
// Private functions
//

func init() {
	var paths = []string{"emptyDir", "multipleEmptyDirs/a", "multipleEmptyDirs/b/c", "multipleEmptyDirs/d/e/f"}

	// Get absolute paths for testdata directories.

	TestdataPathFilesystem, _ = filepath.Abs(".")

	// Kind of hacky.  We want to share testdata/filesystem between internal/filesystem and source, but that means the
	// current directory could be different things depending on which test was launched.  So, try to work around that.

	if strings.HasSuffix(TestdataPathFilesystem, "filesystem") {
		TestdataPathFilesystem, _ = filepath.Abs("../testdata/filesystem")
	} else if strings.HasSuffix(TestdataPathFilesystem, "source") {
		TestdataPathFilesystem, _ = filepath.Abs("../internal/testdata/filesystem")
	} else {
		TestdataPathFilesystem, _ = filepath.Abs("../testdata/filesystem")
	}

	// Initialize empty directories under testdata since we can't store them in Git.

	for _, path := range paths {
		_ = os.MkdirAll(TestdataPathFilesystem+"/"+path, 0775)
	}
}
