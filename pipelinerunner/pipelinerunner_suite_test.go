package pipelinerunner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	apiPrefix = "/api/v1"
)

func TestPipelineRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PipelineRunner Suite")
}
