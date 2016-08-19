package in_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/concourse/atc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api/apifakes"
	"github.com/robdimsdale/concourse-pipeline-resource/in"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/sanitizer"
)

var _ = Describe("In", func() {
	var (
		downloadDir string

		ginkgoLogger logger.Logger

		target   string
		username string
		password string

		inRequest concourse.InRequest
		inCommand *in.InCommand

		fakeAPIClient *apifakes.FakeClient

		pipelines        []api.Pipeline
		pipelineVersions []string

		pipelinesErr      error
		pipelineConfigErr error

		pipelineContents []string
	)

	BeforeEach(func() {
		fakeAPIClient = &apifakes.FakeClient{}

		var err error
		downloadDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		target = "some target"
		username = "some user"
		password = "some password"

		pipelinesErr = nil
		pipelines = []api.Pipeline{
			{
				Name: "pipeline-1",
				URL:  "pipeline_URL_1",
			},
			{
				Name: "pipeline-2",
				URL:  "pipeline_URL_2",
			},
		}

		pipelineVersions = []string{"1234", "2345"}

		pipelineConfigErr = nil
		pipelineContents = make([]string, 2)

		pipelineContents[0] = `---
pipeline1: foo
`

		pipelineContents[1] = `---
pipeline2: foo
`

		inRequest = concourse.InRequest{
			Source: concourse.Source{
				Target:   target,
				Username: username,
				Password: password,
			},
			Version: concourse.Version{
				pipelines[0].Name: pipelineVersions[0],
			},
		}
	})

	JustBeforeEach(func() {
		fakeAPIClient.PipelinesReturns(pipelines, pipelinesErr)

		fakeAPIClient.PipelineConfigStub = func(name string) (atc.Config, string, string, error) {
			defer GinkgoRecover()
			ginkgoLogger.Debugf("GetPipelineStub for: %s\n", name)

			if pipelineConfigErr != nil {
				return atc.Config{}, "", "", pipelineConfigErr
			}

			switch name {
			case pipelines[0].Name:
				return atc.Config{}, pipelineContents[0], pipelineVersions[0], nil
			case pipelines[1].Name:
				return atc.Config{}, pipelineContents[1], pipelineVersions[1], nil
			default:
				Fail("Unexpected invocation of PipelineConfig")
				return atc.Config{}, "", "", nil
			}
		}

		sanitized := concourse.SanitizedSource(inRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		binaryVersion := "v0.1.2-unit-tests"
		inCommand = in.NewInCommand(binaryVersion, ginkgoLogger, fakeAPIClient, downloadDir)
	})

	AfterEach(func() {
		err := os.RemoveAll(downloadDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("downloads all pipeline configs to the target directory", func() {
		_, err := inCommand.Run(inRequest)

		Expect(err).NotTo(HaveOccurred())

		files, err := ioutil.ReadDir(downloadDir)
		Expect(err).NotTo(HaveOccurred())

		Expect(files).To(HaveLen(len(pipelines)))
		Expect(files[0].Name()).To(MatchRegexp("%s.yml", pipelines[0].Name))

		contents, err := ioutil.ReadFile(filepath.Join(downloadDir, files[0].Name()))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal(pipelineContents[0]))

		Expect(files[1].Name()).To(MatchRegexp("%s.yml", pipelines[1].Name))

		contents, err = ioutil.ReadFile(filepath.Join(downloadDir, files[1].Name()))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal(pipelineContents[1]))
	})

	It("returns provided version", func() {
		response, err := inCommand.Run(inRequest)

		Expect(err).NotTo(HaveOccurred())

		Expect(response.Version[pipelines[0].Name]).To(Equal(pipelineVersions[0]))
	})

	It("returns metadata", func() {
		response, err := inCommand.Run(inRequest)

		Expect(err).NotTo(HaveOccurred())

		Expect(response.Metadata).NotTo(BeNil())
	})

	Context("when getting pipelines returns an error", func() {
		BeforeEach(func() {
			pipelinesErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(pipelinesErr))
		})
	})

	Context("when getting pipeline returns an error", func() {
		BeforeEach(func() {
			pipelineConfigErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(Equal(pipelineConfigErr))
		})
	})
})
