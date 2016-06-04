package out_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api/apifakes"
	"github.com/robdimsdale/concourse-pipeline-resource/fly/flyfakes"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/out"
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
		flyRunCallCount    int

		pipelines       []api.Pipeline
		getPipelinesErr error
		setPipelinesErr error

		pipelineConfigErr error
		pipelineContents  []string

		outRequest concourse.OutRequest
		outCommand *out.OutCommand

		fakeFlyConn   *flyfakes.FakeFlyConn
		fakeAPIClient *apifakes.FakeClient
	)

	BeforeEach(func() {
		flyRunCallCount = 0
		fakeFlyConn = &flyfakes.FakeFlyConn{}
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

		fakeAPIClient.PipelineConfigStub = func(name string) (string, error) {
			defer GinkgoRecover()
			ginkgoLogger.Debugf("GetPipelineStub for: %s\n", name)

			if pipelineConfigErr != nil {
				return "", pipelineConfigErr
			}

			switch name {
			case pipelines[0].Name:
				return pipelineContents[0], nil
			case pipelines[1].Name:
				return pipelineContents[1], nil
			case pipelines[2].Name:
				return pipelineContents[2], nil
			default:
				Fail("Unexpected invocation of PipelineConfig")
				return "", nil
			}
		}

		fakeFlyConn.SetPipelineReturns(nil, setPipelinesErr)

		sanitized := concourse.SanitizedSource(outRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		binaryVersion := "v0.1.2-unit-tests"
		outCommand = out.NewOutCommand(binaryVersion, ginkgoLogger, fakeFlyConn, fakeAPIClient, sourcesDir)
	})

	AfterEach(func() {
		err := os.RemoveAll(sourcesDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("invokes fly set-pipeline for each pipeline", func() {
		_, err := outCommand.Run(outRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFlyConn.SetPipelineCallCount()).To(Equal(len(concoursePipelines)))
		for i, p := range concoursePipelines {
			name, configFilepath, varsFilepaths := fakeFlyConn.SetPipelineArgsForCall(i)
			Expect(name).To(Equal(p.Name))
			Expect(configFilepath).To(Equal(filepath.Join(sourcesDir, p.ConfigFile)))

			// the first pipeline has vars files
			if i == 0 {
				Expect(varsFilepaths[0]).To(Equal(filepath.Join(sourcesDir, p.VarsFiles[0])))
				Expect(varsFilepaths[1]).To(Equal(filepath.Join(sourcesDir, p.VarsFiles[1])))
			}
		}
	})

	It("returns provided version", func() {
		response, err := outCommand.Run(outRequest)

		Expect(err).NotTo(HaveOccurred())

		Expect(response.Version.PipelinesChecksum).To(Equal("621f716f112c3c1621bfcfa57dc4f765"))
	})

	It("returns metadata", func() {
		response, err := outCommand.Run(outRequest)

		Expect(err).NotTo(HaveOccurred())

		Expect(response.Metadata).NotTo(BeNil())
	})

	Context("when insecure parses as true", func() {
		BeforeEach(func() {
			outRequest.Source.Insecure = "true"
		})

		It("invokes the login with insecure: true, without error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFlyConn.LoginCallCount()).To(Equal(1))
			_, _, _, insecure := fakeFlyConn.LoginArgsForCall(0)

			Expect(insecure).To(BeTrue())
		})
	})

	Context("when insecure fails to parse into a boolean", func() {
		BeforeEach(func() {
			outRequest.Source.Insecure = "unparsable"
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when login returns an error", func() {
		var (
			expectedErr error
		)

		BeforeEach(func() {
			expectedErr = fmt.Errorf("login failed")
			fakeFlyConn.LoginReturns(nil, expectedErr)
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(expectedErr))
		})
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
