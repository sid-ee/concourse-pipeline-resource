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
	"github.com/robdimsdale/concourse-pipeline-resource/in"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/out/outfakes"
	"github.com/robdimsdale/sanitizer"
)

var _ = Describe("In", func() {
	var (
		downloadDir string

		testLogger in.Logger

		target   string
		username string
		password string

		inRequest concourse.InRequest
		inCommand *in.InCommand

		fakeAPIClient *outfakes.FakeClient

		pipelines        []api.Pipeline
		pipelineVersions []string

		pipelinesErr      error
		pipelineConfigErr error

		pipelineContents []string
	)

	BeforeEach(func() {
		fakeAPIClient = &outfakes.FakeClient{}

		var err error
		downloadDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		target = "some target"
		username = "some user"
		password = "some password"

		pipelinesErr = nil
		pipelines = []api.Pipeline{
			{
				Name:     "pipeline-0",
				TeamName: teamNames[0],
				URL:      "pipeline_URL_0",
			},
			{
				Name:     "pipeline-1",
				TeamName: teamNames[0],
				URL:      "pipeline_URL_1",
			},
			{
				Name:     "pipeline-2",
				TeamName: teamNames[1],
				URL:      "pipeline_URL_2",
			},
		}

		pipelineVersions = []string{"1234", "2345", "3456"}

		pipelineConfigErr = nil
		pipelineContents = make([]string, 3)

		pipelineContents[0] = `---
pipeline0: foo
`

		pipelineContents[1] = `---
pipeline1: foo
`

		pipelineContents[2] = `---
pipeline2: foo
`

		inRequest = concourse.InRequest{
			Source: concourse.Source{
				Target: target,
				Teams: []concourse.Team{
					{
						Name:     teamNames[0],
						Username: username,
						Password: password,
					},
					{
						Name:     teamNames[1],
						Username: username,
						Password: password,
					},
				},
			},
			Version: concourse.Version{
				pipelines[0].Name: pipelineVersions[0],
			},
		}
	})

	JustBeforeEach(func() {
		fakeAPIClient.PipelinesStub = func(teamName string) ([]api.Pipeline, error) {
			switch teamName {
			case teamNames[0]:
				return []api.Pipeline{pipelines[0], pipelines[1]}, pipelinesErr
			case teamNames[1]:
				return []api.Pipeline{pipelines[2]}, pipelinesErr
			default:
				Fail("Unexpected invocation of Pipelines")
				return nil, nil
			}
		}

		fakeAPIClient.PipelineConfigStub = func(teamName string, name string) (atc.Config, string, string, error) {
			defer GinkgoRecover()
			testLogger.Debugf("GetPipelineStub for: %s\n", name)

			if pipelineConfigErr != nil {
				return atc.Config{}, "", "", pipelineConfigErr
			}

			switch name {
			case pipelines[0].Name:
				return atc.Config{}, pipelineContents[0], pipelineVersions[0], nil
			case pipelines[1].Name:
				return atc.Config{}, pipelineContents[1], pipelineVersions[1], nil
			case pipelines[2].Name:
				return atc.Config{}, pipelineContents[2], pipelineVersions[2], nil
			default:
				Fail("Unexpected invocation of PipelineConfig")
				return atc.Config{}, "", "", nil
			}
		}

		sanitized := concourse.SanitizedSource(inRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		testLogger = logger.NewLogger(sanitizer)

		binaryVersion := "v0.1.2-unit-tests"
		inCommand = in.NewInCommand(binaryVersion, testLogger, fakeAPIClient, downloadDir)
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

		for i, pipeline := range pipelines {
			Expect(files[i].Name()).To(MatchRegexp("%s-%s.yml", pipeline.TeamName, pipeline.Name))
			contents, err := ioutil.ReadFile(filepath.Join(downloadDir, files[i].Name()))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal(pipelineContents[i]))
		}
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
