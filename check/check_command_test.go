package check_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/robdimsdale/concourse-pipeline-resource/check"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/fly/flyfakes"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/sanitizer"
)

var _ = Describe("Check", func() {
	var (
		server *ghttp.Server

		tempDir     string
		logFilePath string

		ginkgoLogger logger.Logger

		target   string
		username string
		password string

		pipelinesChecksum string
		expectedResponse  concourse.CheckResponse
		pipelineContents  []string

		checkRequest concourse.CheckRequest
		checkCommand *check.CheckCommand

		pipelinesResponseStatusCode *int
		pipelinesResponseBody       *[]concourse.Pipeline
		fakeFlyConn                 *flyfakes.FakeFlyConn
		runCallCount                int
	)

	BeforeEach(func() {
		runCallCount = 0
		fakeFlyConn = &flyfakes.FakeFlyConn{}

		server = ghttp.NewServer()

		pipelinesResponseBody = &[]concourse.Pipeline{
			{
				Name: "pipeline 1",
				URL:  "pipeline URL 1",
			},
			{
				Name: "pipeline 2",
				URL:  "pipeline URL 2",
			},
		}

		pipelineContents = make([]string, 2)

		pipelineContents[0] = `---
pipeline1: foo
`

		pipelineContents[1] = `---
pipeline2: foo
`

		fakeFlyConn.RunStub = func(...string) ([]byte, error) {
			switch runCallCount {
			case 0:
				return []byte(pipelineContents[0]), nil
			case 1:
				return []byte(pipelineContents[1]), nil
			default:
				Fail("Unexpected invocation of flyConn.Run")
			}
			runCallCount += 1
			return nil, nil
		}

		pipelinesResponseStatusCode = new(int)
		*pipelinesResponseStatusCode = http.StatusOK
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", fmt.Sprintf(
					"%s/pipelines",
					apiPrefix,
				)),
				ghttp.RespondWithJSONEncodedPtr(
					pipelinesResponseStatusCode,
					pipelinesResponseBody,
				),
			),
		)

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		logFilePath = filepath.Join(tempDir, "concourse-pipeline-resource-check.log1234")
		err = ioutil.WriteFile(logFilePath, []byte("initial log content"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		binaryVersion := "v0.1.2-unit-tests"

		target = server.URL()
		username = "some user"
		password = "some password"

		checkRequest = concourse.CheckRequest{
			Source: concourse.Source{
				Target:   target,
				Username: username,
				Password: password,
			},
		}

		sanitized := concourse.SanitizedSource(checkRequest.Source)
		sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)

		ginkgoLogger = logger.NewLogger(sanitizer)

		pipelinesChecksum = "2e28dea4f7ce0c811f3035cdb831d74b"
		expectedResponse = []concourse.Version{
			{
				PipelinesChecksum: pipelinesChecksum,
			},
		}

		checkCommand = check.NewCheckCommand(
			binaryVersion,
			ginkgoLogger,
			logFilePath,
			fakeFlyConn,
		)
	})

	AfterEach(func() {
		server.Close()

		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns pipelines checksum without error", func() {
		response, err := checkCommand.Run(checkRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(response).To(Equal(expectedResponse))
	})

	Context("when no target is provided", func() {
		BeforeEach(func() {
			checkRequest.Source.Target = ""
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*target.*provided"))
		})
	})

	Context("when no username is provided", func() {
		BeforeEach(func() {
			checkRequest.Source.Username = ""
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*username.*provided"))
		})
	})

	Context("when no password is provided", func() {
		BeforeEach(func() {
			checkRequest.Source.Password = ""
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(MatchRegexp(".*password.*provided"))
		})
	})

	Context("when log files already exist", func() {
		var (
			otherFilePath1 string
			otherFilePath2 string
		)

		BeforeEach(func() {
			otherFilePath1 = filepath.Join(tempDir, "concourse-pipeline-resource-check.log1")
			err := ioutil.WriteFile(otherFilePath1, []byte("initial log content"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			otherFilePath2 = filepath.Join(tempDir, "concourse-pipeline-resource-check.log2")
			err = ioutil.WriteFile(otherFilePath2, []byte("initial log content"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("removes the other log files", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(otherFilePath1)
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())

			_, err = os.Stat(otherFilePath2)
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())

			_, err = os.Stat(logFilePath)
			Expect(err).NotTo(HaveOccurred())
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
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(expectedErr))
		})
	})

	Context("when getting pipelines returns an error", func() {
		BeforeEach(func() {
			checkRequest.Source.Target = "some-bad-target"
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when getting pipelines returns a non-expected status code", func() {
		BeforeEach(func() {
			*pipelinesResponseStatusCode = http.StatusForbidden
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when getting pipelines returns unmarshallable body", func() {
		BeforeEach(func() {
			server.Reset()

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/pipelines",
						apiPrefix,
					)),
					ghttp.RespondWith(
						http.StatusOK,
						`$not%valid-#json`,
					),
				),
			)
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when calling fly to get pipeline config returns an error", func() {
		var (
			expectedErr error
		)

		BeforeEach(func() {
			expectedErr = fmt.Errorf("error executing fly")

			fakeFlyConn.RunStub = func(...string) ([]byte, error) {
				return nil, expectedErr
			}
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(expectedErr))
		})
	})

})
