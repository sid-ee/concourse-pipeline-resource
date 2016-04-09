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

		target          string
		username        string
		password        string
		pipelines       []concourse.Pipeline
		flyRunCallCount int

		apiPipelines []api.Pipeline
		pipelinesErr error

		pipelineContents []string

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

		apiPipelines = []api.Pipeline{
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
		pipelinesErr = nil

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

		pipelines = []concourse.Pipeline{
			{
				Name:       apiPipelines[0].Name,
				ConfigFile: "pipeline_1.yml",
				VarsFiles: []string{
					"vars_1.yml",
					"vars_2.yml",
				},
			},
			{
				Name:       apiPipelines[1].Name,
				ConfigFile: "pipeline_2.yml",
			},
		}

		fakeFlyConn.RunStub = func(args ...string) ([]byte, error) {
			defer GinkgoRecover()

			switch args[0] {
			case "get-pipeline":
				// args[1] will be "-p"
				switch args[2] {
				case apiPipelines[0].Name:
					return []byte(pipelineContents[0]), nil
				case apiPipelines[1].Name:
					return []byte(pipelineContents[1]), nil
				case apiPipelines[2].Name:
					return []byte(pipelineContents[2]), nil
				}

			case "set-pipeline":
				return nil, nil

			default:
				Fail(fmt.Sprintf("Unexpected invocation of flyConn.Run: %+v", args))
			}
			return nil, nil
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
	})

	JustBeforeEach(func() {
		fakeAPIClient.PipelinesReturns(apiPipelines, pipelinesErr)

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

		// There may be some later calls to fly (e.g. getting the pipeline)
		Expect(fakeFlyConn.RunCallCount()).To(BeNumerically(">", len(pipelines)))
		for i, p := range pipelines {
			args := fakeFlyConn.RunArgsForCall(i)
			Expect(args[0]).To(Equal("set-pipeline"))
			Expect(args[1]).To(Equal("-n"))
			Expect(args[2]).To(Equal("-p"))
			Expect(args[3]).To(Equal(p.Name))
			Expect(args[4]).To(Equal("-c"))
			Expect(args[5]).To(Equal(filepath.Join(sourcesDir, p.ConfigFile)))

			// the first pipeline has vars files
			if i == 0 {
				Expect(args[6]).To(Equal("-l"))
				Expect(args[7]).To(Equal(filepath.Join(sourcesDir, p.VarsFiles[0])))
				Expect(args[8]).To(Equal("-l"))
				Expect(args[9]).To(Equal(filepath.Join(sourcesDir, p.VarsFiles[1])))
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

	Context("when getting pipelines returns an error", func() {
		BeforeEach(func() {
			pipelinesErr = fmt.Errorf("some error")
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(pipelinesErr))
		})
	})

	Context("when getting pipeline returns an error", func() {
		var (
			expectedErr error
		)

		BeforeEach(func() {
			expectedErr = fmt.Errorf("some error")

			fakeFlyConn.RunStub = func(args ...string) ([]byte, error) {
				defer GinkgoRecover()

				switch args[0] {
				case "get-pipeline":
					return nil, expectedErr

				case "set-pipeline":
					return nil, nil

				default:
					Fail(fmt.Sprintf("Unexpected invocation of flyConn.Run: %+v", args))
				}
				return nil, nil
			}
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(expectedErr))
		})
	})
})
