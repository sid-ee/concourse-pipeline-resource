package in_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/fly/flyfakes"
	"github.com/robdimsdale/concourse-pipeline-resource/in"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/sanitizer"
)

var _ = Describe("In", func() {
	var (
		server *ghttp.Server

		downloadDir string

		ginkgoLogger logger.Logger

		target   string
		username string
		password string

		flyBinaryPath string

		inRequest concourse.InRequest
		inCommand *in.InCommand

		fakeFlyConn     *flyfakes.FakeFlyConn
		flyRunCallCount int

		pipelinesResponseStatusCode int
		pipelines                   []api.Pipeline

		pipelineResponseStatusCode int

		pipelineContents []string
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		fakeFlyConn = &flyfakes.FakeFlyConn{}
		flyRunCallCount = 0

		var err error
		downloadDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		target = server.URL()
		username = "some user"
		password = "some password"
		flyBinaryPath = "fly"

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
		pipelinesResponseStatusCode = http.StatusOK
		pipelineResponseStatusCode = http.StatusOK

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
		}

		fakeFlyConn.RunStub = func(args ...string) ([]byte, error) {
			if args[0] == "gp" {
				switch args[2] {
				case pipelines[0].Name:
					return []byte(pipelineContents[0]), nil
				case pipelines[1].Name:
					return []byte(pipelineContents[1]), nil
				}
			}

			flyRunCallCount += 1
			switch flyRunCallCount {
			case 1:
				return []byte(pipelineContents[0]), nil
			case 2:
				return []byte(pipelineContents[1]), nil
			default:
				Fail("Unexpected invocation of flyConn.Run")
			}
			return nil, nil
		}
	})

	JustBeforeEach(func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", fmt.Sprintf(
					"%s/pipelines",
					apiPrefix,
				)),
				ghttp.RespondWithJSONEncoded(
					pipelinesResponseStatusCode,
					pipelines,
				),
			),
		)

		for i, p := range pipelines {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/pipelines/%d",
						apiPrefix,
						p.URL,
					)),
					ghttp.RespondWithJSONEncoded(
						pipelineResponseStatusCode,
						pipelines[i],
					),
				),
			)
		}

		sanitized := concourse.SanitizedSource(inRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		binaryVersion := "v0.1.2-unit-tests"
		inCommand = in.NewInCommand(binaryVersion, ginkgoLogger, fakeFlyConn, downloadDir)
	})

	AfterEach(func() {
		server.Close()

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

	Context("when login returns an error", func() {
		var (
			expectedErr error
		)

		BeforeEach(func() {
			expectedErr = fmt.Errorf("login failed")
			fakeFlyConn.LoginReturns(nil, expectedErr)
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(expectedErr))
		})
	})

	Context("when getting pipelines returns an error", func() {
		BeforeEach(func() {
			inRequest.Source.Target = "some-bad-target"
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when getting pipeline returns an error", func() {
		var (
			expectedErr error
		)

		BeforeEach(func() {
			expectedErr = fmt.Errorf("some error")
			fakeFlyConn.RunReturns(nil, expectedErr)
		})

		It("returns an error", func() {
			_, err := inCommand.Run(inRequest)
			Expect(err).To(Equal(expectedErr))
		})
	})
})
