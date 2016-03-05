package in_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/in"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/sanitizer"
)

var _ = Describe("In", func() {
	var (
		server *ghttp.Server

		tempDir string

		ginkgoLogger logger.Logger

		target   string
		username string
		password string

		flyBinaryPath string

		inRequest concourse.InRequest
		inCommand *in.InCommand
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		binaryVersion := "v0.1.2-unit-tests"

		target = server.URL()
		username = "some user"
		password = "some password"

		inRequest = concourse.InRequest{
			Source: concourse.Source{
				Target:   target,
				Username: username,
				Password: password,
			},
		}

		sanitized := concourse.SanitizedSource(inRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		flyBinaryPath = "fly"
		inCommand = in.NewInCommand(binaryVersion, ginkgoLogger, flyBinaryPath)
	})

	AfterEach(func() {
		server.Close()

		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when no target is provided", func() {
		BeforeEach(func() {
			inRequest.Source.Target = ""
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*target.*provided"))
		})
	})

	Context("when no username is provided", func() {
		BeforeEach(func() {
			inRequest.Source.Username = ""
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*username.*provided"))
		})
	})

	Context("when no password is provided", func() {
		BeforeEach(func() {
			inRequest.Source.Password = ""
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*password.*provided"))
		})
	})
})
