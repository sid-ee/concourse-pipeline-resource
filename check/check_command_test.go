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

		pipelinesErr      error
		pipelineConfigErr error
		pipelines         []api.Pipeline

		fakeAPIClient *apifakes.FakeClient
	)

	BeforeEach(func() {
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

		pipelineConfigErr = nil

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
			fakeAPIClient,
		)
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		fakeAPIClient.PipelinesReturns(pipelines, pipelinesErr)

		fakeAPIClient.PipelineConfigStub = func(name string) (string, error) {
			ginkgoLogger.Debugf("GetPipelineStub for: %s\n", name)

			if pipelineConfigErr != nil {
				return "", pipelineConfigErr
			}

			switch name {
			case pipelines[0].Name:
				return pipelineContents[0], nil
			case pipelines[1].Name:
				return pipelineContents[1], nil
			default:
				Fail("Unexpected invocation of flyConn.GetPipeline")
				return "", nil
			}
		}
	})

	It("returns pipelines checksum without error", func() {
		response, err := checkCommand.Run(checkRequest)
		Expect(err).NotTo(HaveOccurred())

		Expect(response).To(Equal(expectedResponse))
	})

	Context("when the most recent version is provided", func() {
		BeforeEach(func() {
			checkRequest.Version = concourse.Version{
				PipelinesChecksum: expectedResponse[0].PipelinesChecksum,
			}
		})

		It("returns the most recent version", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(Equal(expectedResponse))
		})
	})

	Context("when some other version is provided", func() {
		BeforeEach(func() {
			checkRequest.Version = concourse.Version{
				PipelinesChecksum: "foo",
			}
		})

		It("returns the most recent version", func() {
			response, err := checkCommand.Run(checkRequest)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(Equal(expectedResponse))
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

	Context("when getting pipeline config returns an error", func() {
		BeforeEach(func() {
			pipelineConfigErr = fmt.Errorf("error getting pipelineConfig")
		})

		It("returns an error", func() {
			_, err := checkCommand.Run(checkRequest)
			Expect(err).To(HaveOccurred())

			Expect(err).To(Equal(pipelineConfigErr))
		})
	})
})
