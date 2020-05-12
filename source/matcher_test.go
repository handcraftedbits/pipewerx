package source // import "golang.handcraftedbits.com/pipewerx/source"

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"golang.handcraftedbits.com/pipewerx"
)

//
// Private types
//

type matcherHaveTheseFiles struct {
	contents  []string
	paths     []string
	separator string
}

func (matcher *matcherHaveTheseFiles) Match(actual interface{}) (bool, error) {
	var contentMap = make(map[string]string)
	var ok bool
	var pathMap = make(map[string]bool)
	var results []pipewerx.Result

	results, ok = actual.([]pipewerx.Result)

	if !ok {
		return false, errors.New("haveTheseFiles expected []pipewerx.Result")
	}

	for i, path := range matcher.paths {
		// For simplicity's sake we're using '/' as a path separator, but that might not be true for all filesystems.
		// To address that, replace all '/' with the correct path separator before adding to the map.

		path = strings.ReplaceAll(path, "/", matcher.separator)

		pathMap[path] = true

		if matcher.contents != nil {
			contentMap[path] = matcher.contents[i]
		}
	}

	for _, result := range results {
		Expect(result.Error()).To(BeNil())
		Expect(pathMap).To(HaveKey(result.File().Path().String()))

		if matcher.contents != nil {
			var err error
			var fileContents []byte
			var reader io.ReadCloser

			Expect(contentMap).To(HaveKey(result.File().Path().String()))

			reader, err = result.File().Reader()

			Expect(err).To(BeNil())
			Expect(reader).NotTo(BeNil())

			fileContents, err = ioutil.ReadAll(reader)

			Expect(err).To(BeNil())
			Expect(fileContents).NotTo(BeNil())

			err = reader.Close()

			Expect(err).To(BeNil())
			Expect(result.File().Size()).To(Equal(int64(len(fileContents))))
			Expect(string(fileContents)).To(Equal(contentMap[result.File().Path().String()]))
		}
	}

	return true, nil
}

func (matcher *matcherHaveTheseFiles) FailureMessage(actual interface{}) string {
	return ""
}

func (matcher *matcherHaveTheseFiles) NegatedFailureMessage(actual interface{}) string {
	return ""
}

//
// Private functions
//

func haveTheseFiles(separator string, paths []string, contents []string) types.GomegaMatcher {
	return &matcherHaveTheseFiles{
		contents:  contents,
		paths:     paths,
		separator: separator,
	}
}
