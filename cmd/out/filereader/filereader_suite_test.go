package filereader_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	apiPrefix = "/api/v1"
)

func TestOutFilereader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OutFilereader Suite")
}
