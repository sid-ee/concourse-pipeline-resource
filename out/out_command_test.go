package out_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/concourse/concourse-pipeline-resource/concourse"
	"github.com/concourse/concourse-pipeline-resource/concourse/api"
	"github.com/concourse/concourse-pipeline-resource/concourse/api/apifakes"
	"github.com/concourse/concourse-pipeline-resource/fly/flyfakes"
	"github.com/concourse/concourse-pipeline-resource/logger"
	"github.com/concourse/concourse-pipeline-resource/out"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/sanitizer"
)

var _ = Describe("Out", func() {
	var (
		sourcesDir string

		ginkgoLogger logger.Logger

		target        string
		username      string
		otherUsername string
		password      string
		otherPassword string
		teamName      string
		otherTeamName string
		pipelines     []concourse.Pipeline

		apiPipelines    []api.Pipeline
		getPipelinesErr error
		setPipelinesErr error

		pipelineContents []string

		outRequest    concourse.OutRequest
		badOutRequest concourse.OutRequest
		outCommand    *out.OutCommand

		fakeFlyConn   *flyfakes.FakeFlyConn
		fakeAPIClient *apifakes.FakeClient
	)

	BeforeEach(func() {
		fakeFlyConn = &flyfakes.FakeFlyConn{}
		fakeAPIClient = &apifakes.FakeClient{}

		var err error
		sourcesDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		target = "some target"
		username = "some user"
		otherUsername = "some other user"
		password = "some password"
		otherPassword = "some other password"
		teamName = "main"
		otherTeamName = "some-other-team"

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

		pipelines = []concourse.Pipeline{
			{
				Name:       apiPipelines[0].Name,
				ConfigFile: "pipeline_1.yml",
				VarsFiles: []string{
					"vars_1.yml",
					"vars_2.yml",
				},
				TeamName: teamName,
			},
			{
				Name:       apiPipelines[1].Name,
				ConfigFile: "pipeline_2.yml",
				TeamName:   teamName,
			},
			{
				Name:       apiPipelines[2].Name,
				ConfigFile: "pipeline_3.yml",
				TeamName:   otherTeamName,
			},
		}

		fakeFlyConn.GetPipelineStub = func(name string) ([]byte, error) {
			defer GinkgoRecover()
			ginkgoLogger.Debugf("GetPipelineStub for: %s\n", name)

			switch name {
			case apiPipelines[0].Name:
				return []byte(pipelineContents[0]), nil
			case apiPipelines[1].Name:
				return []byte(pipelineContents[1]), nil
			case apiPipelines[2].Name:
				return []byte(pipelineContents[2]), nil
			default:
				Fail("Unexpected invocation of flyConn.GetPipeline")
				return nil, nil
			}
		}

		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				Target: target,
				Teams: []concourse.Team{
					{
						Name:     teamName,
						Username: username,
						Password: password,
					},
					{
						Name:     otherTeamName,
						Username: otherUsername,
						Password: otherPassword,
					},
				},
			},
			Params: concourse.OutParams{
				Pipelines: pipelines,
			},
		}

		badOutRequest = concourse.OutRequest{
			Source: concourse.Source{
				Target: target,
				Teams: []concourse.Team{
					{
						Name:     "the-wrong-team",
						Username: "the wrong username",
						Password: "the wrong password",
					},
				},
			},
			Params: concourse.OutParams{
				Pipelines: pipelines,
			},
		}
	})

	JustBeforeEach(func() {
		fakeAPIClient.PipelinesReturns(apiPipelines, getPipelinesErr)
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

	It("syncs the fly version to the given target", func() {
		_, err := outCommand.Run(outRequest)
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeFlyConn.SyncCallCount()).To(Equal(1))
	})

	It("invokes fly set-pipeline for each pipeline", func() {
		_, err := outCommand.Run(outRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeFlyConn.SetPipelineCallCount()).To(Equal(len(pipelines)))

		for i, p := range pipelines {
			name, configFilepath, varsFilepaths := fakeFlyConn.SetPipelineArgsForCall(i)
			_, tname, _, _, _ := fakeFlyConn.LoginArgsForCall(i)
			Expect(name).To(Equal(p.Name))
			Expect(tname).To(Equal(p.TeamName))
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

		Expect(response.Version[apiPipelines[0].Name]).To(Equal("4f4bd60b18bf697cc68dac9cb95537c2"))
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

			Expect(fakeFlyConn.LoginCallCount()).To(Equal(5))
			_, _, _, _, insecure := fakeFlyConn.LoginArgsForCall(0)

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

	Context("when setting a pipeline that belongs to another team", func() {
		It("returns an error", func() {
			_, err := outCommand.Run(badOutRequest)
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
		var (
			expectedErr error
		)

		BeforeEach(func() {
			expectedErr = fmt.Errorf("some error")

			fakeFlyConn.GetPipelineReturns(nil, expectedErr)
		})

		It("returns an error", func() {
			_, err := outCommand.Run(outRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(expectedErr))
		})
	})
})
