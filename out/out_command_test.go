package out_test

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
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/out"
	"github.com/robdimsdale/concourse-pipeline-resource/out/helpers/helpersfakes"
	"github.com/robdimsdale/concourse-pipeline-resource/sanitizer"
)

var _ = Describe("Out", func() {
	var (
		sourcesDir string

		ginkgoLogger logger.Logger

		target             string
		username           string
		password           string
		concoursePipelines []concourse.Pipeline

		pipelines       []api.Pipeline
		getPipelinesErr error
		setPipelinesErr error

		pipelineConfigErr error
		pipelineContents  []string
		pipelineVersions  []string

		outRequest concourse.OutRequest
		outCommand *out.OutCommand

		fakePipelineSetter *helpersfakes.FakePipelineSetter
		fakeAPIClient      *apifakes.FakeClient
	)

	BeforeEach(func() {
		fakePipelineSetter = &helpersfakes.FakePipelineSetter{}
		fakeAPIClient = &apifakes.FakeClient{}

		var err error
		sourcesDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		target = "some target"
		username = "some user"
		password = "some password"

		pipelines = []api.Pipeline{
			{
				Name: "pipeline-1",
				URL:  "pipeline_URL_1",
			},
			{
				Name: "pipeline-2",
				URL:  "pipeline_URL_2",
			},
			{
				Name: "pipeline-3",
				URL:  "pipeline_URL_3",
			},
		}
		getPipelinesErr = nil
		setPipelinesErr = nil

		pipelineContents = make([]string, 3)

		pipelineContents[0] = `---
pipeline1: foo
`

		pipelineContents[1] = `---
pipeline2: foo
`

		pipelineContents[2] = `---
pipeline3: foo
`

		pipelineVersions = []string{"1234", "2345", "3456"}

		concoursePipelines = []concourse.Pipeline{
			{
				Name:       pipelines[0].Name,
				ConfigFile: "pipeline_1.yml",
				VarsFiles: []string{
					"vars_1.yml",
					"vars_2.yml",
				},
			},
			{
				Name:       pipelines[1].Name,
				ConfigFile: "pipeline_2.yml",
			},
		}

		pipelineConfigErr = nil

		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				Target:   target,
				Username: username,
				Password: password,
			},
			Params: concourse.OutParams{
				Pipelines: concoursePipelines,
			},
		}
	})

	JustBeforeEach(func() {
		fakeAPIClient.PipelinesReturns(pipelines, getPipelinesErr)

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
			case pipelines[2].Name:
				return atc.Config{}, pipelineContents[2], pipelineVersions[2], nil
			default:
				Fail("Unexpected invocation of PipelineConfig")
				return atc.Config{}, "", "", nil
			}
		}

		fakePipelineSetter.SetPipelineReturns(setPipelinesErr)

		sanitized := concourse.SanitizedSource(outRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		binaryVersion := "v0.1.2-unit-tests"
		outCommand = out.NewOutCommand(
			binaryVersion,
			ginkgoLogger,
			fakePipelineSetter,
			fakeAPIClient,
			sourcesDir,
		)
	})

	AfterEach(func() {
		err := os.RemoveAll(sourcesDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("sets each pipeline", func() {
		_, err := outCommand.Run(outRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakePipelineSetter.SetPipelineCallCount()).To(Equal(len(concoursePipelines)))
		for i, p := range concoursePipelines {
			name, configFilepath, _, varsFilepaths := fakePipelineSetter.SetPipelineArgsForCall(i)
			Expect(name).To(Equal(p.Name))
			Expect(configFilepath).To(Equal(filepath.Join(sourcesDir, p.ConfigFile)))

			// the first pipeline has vars files
			if i == 0 {
				Expect(varsFilepaths[0]).To(Equal(filepath.Join(sourcesDir, p.VarsFiles[0])))
				Expect(varsFilepaths[1]).To(Equal(filepath.Join(sourcesDir, p.VarsFiles[1])))
			}
		}
	})

	It("returns updated pipeline version", func() {
		response, err := outCommand.Run(outRequest)

		Expect(err).NotTo(HaveOccurred())

		Expect(response.Version[pipelines[0].Name]).
			To(Equal(pipelineVersions[0]))
		Expect(response.Version[pipelines[1].Name]).
			To(Equal(pipelineVersions[1]))
		Expect(response.Version[pipelines[2].Name]).
			To(Equal(pipelineVersions[2]))
	})

	It("returns metadata", func() {
		response, err := outCommand.Run(outRequest)

		Expect(err).NotTo(HaveOccurred())

		Expect(response.Metadata).NotTo(BeNil())
	})

	Context("when setting pipelines returns an error", func() {
		BeforeEach(func() {
			setPipelinesErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(setPipelinesErr))
		})
	})

	Context("when getting pipelines returns an error", func() {
		BeforeEach(func() {
			getPipelinesErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(getPipelinesErr))
		})
	})

	Context("when getting pipeline returns an error", func() {
		BeforeEach(func() {
			pipelineConfigErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(pipelineConfigErr))
		})
	})
})
