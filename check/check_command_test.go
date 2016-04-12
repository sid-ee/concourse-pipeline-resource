package check_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/check"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api/apifakes"
	"github.com/robdimsdale/concourse-pipeline-resource/fly/flyfakes"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/sanitizer"
)

var _ = Describe("Check", func() {
	var (
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

		pipelinesErr error
		pipelines    []api.Pipeline
		fakeFlyConn  *flyfakes.FakeFlyConn
		runCallCount int

		fakeAPIClient *apifakes.FakeClient
	)

	BeforeEach(func() {
		runCallCount = 0
		fakeFlyConn = &flyfakes.FakeFlyConn{}
		fakeAPIClient = &apifakes.FakeClient{}

		pipelinesErr = nil
		pipelines = []api.Pipeline{
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

		fakeFlyConn.GetPipelineStub = func(name string) ([]byte, error) {
			ginkgoLogger.Debugf("GetPipelineStub for: %s\n", name)

			switch name {
			case pipelines[0].Name:
				return []byte(pipelineContents[0]), nil
			case pipelines[1].Name:
				return []byte(pipelineContents[1]), nil
			default:
				Fail("Unexpected invocation of flyConn.GetPipeline")
				return nil, nil
			}
		}

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		logFilePath = filepath.Join(tempDir, "concourse-pipeline-resource-check.log1234")
		err = ioutil.WriteFile(logFilePath, []byte("initial log content"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		binaryVersion := "v0.1.2-unit-tests"

		target = "some target"
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

		pipelinesChecksum = "15c2eae5189cedf18c1d7e278b14342a"
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
			fakeAPIClient,
		)
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		fakeAPIClient.PipelinesReturns(pipelines, pipelinesErr)
	})

	It("returns pipelines checksum without error", func() {
		response, err := checkCommand.Run(checkRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(response).To(Equal(expectedResponse))
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
			pipelinesErr = fmt.Errorf("some error")
		})

		It("forwards the error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(pipelinesErr))
		})
	})

	Context("when calling fly to get pipeline config returns an error", func() {
		var (
			expectedErr error
		)

		BeforeEach(func() {
			expectedErr = fmt.Errorf("error executing fly")

			fakeFlyConn.GetPipelineReturns(nil, expectedErr)
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(expectedErr))
		})
	})
})
