package out_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/out"
	"github.com/robdimsdale/concourse-pipeline-resource/sanitizer"
)

var _ = Describe("Out", func() {
	var (
		server *ghttp.Server

		tempDir string

		ginkgoLogger logger.Logger

		target    string
		username  string
		password  string
		pipelines []concourse.Pipeline

		flyBinaryPath string

		outRequest concourse.OutRequest
		outCommand *out.OutCommand
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

		pipelines = []concourse.Pipeline{
			{
				Name:       "pipeline-1",
				ConfigFile: "pipeline_1.yml",
			},
		}

		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				Target:   target,
				Username: username,
				Password: password,
			},
			Params: concourse.OutParams{
				Pipelines: pipelines,
			},
		}

		sanitized := concourse.SanitizedSource(outRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		flyBinaryPath = "fly"
		outCommand = out.NewOutCommand(binaryVersion, ginkgoLogger, flyBinaryPath)
	})

	AfterEach(func() {
		server.Close()

		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when no target is provided", func() {
		BeforeEach(func() {
			outRequest.Source.Target = ""
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*target.*provided"))
		})
	})

	Context("when no username is provided", func() {
		BeforeEach(func() {
			outRequest.Source.Username = ""
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*username.*provided"))
		})
	})

	Context("when no password is provided", func() {
		BeforeEach(func() {
			outRequest.Source.Password = ""
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*password.*provided"))
		})
	})

	Context("when pipelines param is nil", func() {
		BeforeEach(func() {
			outRequest.Params.Pipelines = nil
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*pipelines.*provided"))
		})
	})

	Context("when pipelines param is empty", func() {
		BeforeEach(func() {
			outRequest.Params.Pipelines = []concourse.Pipeline{}
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*pipelines.*provided"))
		})
	})
})
