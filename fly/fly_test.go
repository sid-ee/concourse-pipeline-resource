package fly_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger/loggerfakes"
	"github.com/robdimsdale/concourse-pipeline-resource/fly"
)

var _ = Describe("FlyConn", func() {
	var (
		flyConn fly.FlyConn

		target        string
		flyBinaryPath string

		fakeLogger *loggerfakes.FakeLogger
	)

	BeforeEach(func() {
		target = "some-target"
		flyBinaryPath = "echo"

		fakeLogger = &loggerfakes.FakeLogger{}

		flyConn = fly.NewFlyConn(target, fakeLogger, flyBinaryPath)
	})

	Describe("Login", func() {
		var (
			url      string
			username string
			password string
		)

		BeforeEach(func() {
			url = "some-url"
			username = "some-username"
			password = "some-password"
		})

		It("returns output without error", func() {
			output, _, err := flyConn.Login(url, username, password)
			Expect(err).NotTo(HaveOccurred())

			expectedOutput := fmt.Sprintf(
				"%s %s %s %s %s %s %s %s %s\n",
				"-t", target,
				"login",
				"-c", url,
				"-u", username,
				"-p", password,
			)

			Expect(string(output)).To(Equal(expectedOutput))
		})
	})

	Describe("GetPipeline", func() {
		var (
			pipelineName string
		)

		BeforeEach(func() {
			pipelineName = "some-pipeline"
		})

		It("returns output without error", func() {
			output, _, err := flyConn.GetPipeline(pipelineName)
			Expect(err).NotTo(HaveOccurred())

			expectedOutput := fmt.Sprintf(
				"%s %s %s %s %s\n",
				"-t", target,
				"get-pipeline",
				"-p", pipelineName,
			)

			Expect(string(output)).To(Equal(expectedOutput))
		})
	})

	Describe("SetPipeline", func() {
		var (
			pipelineName   string
			configFilepath string
		)

		BeforeEach(func() {
			pipelineName = "some-pipeline"
			configFilepath = "some-config-file"
		})

		It("returns output without error", func() {
			output, _, err := flyConn.SetPipeline(pipelineName, configFilepath, nil)
			Expect(err).NotTo(HaveOccurred())

			expectedOutput := fmt.Sprintf(
				"%s %s %s %s %s %s %s %s\n",
				"-t", target,
				"set-pipeline",
				"-n",
				"-p", pipelineName,
				"-c", configFilepath,
			)

			Expect(string(output)).To(Equal(expectedOutput))
		})

		Context("when optional vars files are provided", func() {

			var (
				varsFiles []string
			)

			BeforeEach(func() {
				varsFiles = []string{
					"vars-file-1",
					"vars-file-2",
				}
			})

			It("returns output without error", func() {
				output, _, err := flyConn.SetPipeline(pipelineName, configFilepath, varsFiles)
				Expect(err).NotTo(HaveOccurred())

				expectedOutput := fmt.Sprintf(
					"%s %s %s %s %s %s %s %s %s %s %s %s\n",
					"-t", target,
					"set-pipeline",
					"-n",
					"-p", pipelineName,
					"-c", configFilepath,
					"-l", varsFiles[0],
					"-l", varsFiles[1],
				)

				Expect(string(output)).To(Equal(expectedOutput))
			})
		})
	})
})
