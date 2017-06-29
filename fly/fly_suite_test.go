package fly_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	apiPrefix = "/api/v1"
)

func TestFly(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fly Suite")
}
