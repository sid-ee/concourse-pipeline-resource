package fly_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/concourse-pipeline-resource/fly"
	"github.com/robdimsdale/concourse-pipeline-resource/logger/loggerfakes"
)

const (
	errScript = `#!/bin/sh
>&1 echo "some std output"
>&2 echo "some err output"
exit 1
`
)

var _ = Describe("FlyConn", func() {
	var (
		flyConn fly.FlyConn

		target string

		tempDir         string
		flyBinaryPath   string
		fakeFlyContents string

		fakeLogger *loggerfakes.FakeLogger
	)

	BeforeEach(func() {
		target = "some-target"

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		flyBinaryPath = filepath.Join(tempDir, "fake_fly")

		fakeFlyContents = `#!/bin/sh
		echo $@`

		fakeLogger = &loggerfakes.FakeLogger{}
	})

	JustBeforeEach(func() {
		err := ioutil.WriteFile(flyBinaryPath, []byte(fakeFlyContents), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		flyConn = fly.NewFlyConn(target, fakeLogger, flyBinaryPath)
	})

	AfterEach(func() {
		err := os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Login", func() {
		var (
			url      string
			username string
			password string
			insecure string
		)

		BeforeEach(func() {
			url = "some-url"
			username = "some-username"
			password = "some-password"
			insecure = "false"
		})

		It("returns output without error", func() {
			output, err := flyConn.Login(url, username, password,insecure)
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
		Context("when insecure login is enabled", func() {
			BeforeEach(func() {
				insecure="true"
			})
			It("returns output without error", func() {
				output, err := flyConn.Login(url, username, password,insecure)
				Expect(err).NotTo(HaveOccurred())

				expectedOutput := fmt.Sprintf(
					"%s %s %s %s %s %s %s %s %s %s\n",
					"-t", target,
					"login",
					"-c", url,
					"-u", username,
					"-p", password,
					"-k",
				)

				Expect(string(output)).To(Equal(expectedOutput))
			})

		})

		Context("when there is an error starting the commmand", func() {
			BeforeEach(func() {
				fakeFlyContents = ""
			})

			It("returns an error", func() {
				_, err := flyConn.Login(url, username, password, insecure)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the command returns an error", func() {
			BeforeEach(func() {
				fakeFlyContents = errScript
			})

			It("appends stderr to the error", func() {
				_, err := flyConn.Login(url, username, password, insecure)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(MatchRegexp(".*some err output.*"))
			})
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
			output, err := flyConn.GetPipeline(pipelineName)
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
			output, err := flyConn.SetPipeline(pipelineName, configFilepath, nil)
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
				output, err := flyConn.SetPipeline(pipelineName, configFilepath, varsFiles)
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

	Describe("DestroyPipeline", func() {
		var (
			pipelineName   string
			configFilepath string
		)

		BeforeEach(func() {
			pipelineName = "some-pipeline"
			configFilepath = "some-config-file"
		})

		It("returns output without error", func() {
			output, err := flyConn.DestroyPipeline(pipelineName)
			Expect(err).NotTo(HaveOccurred())

			expectedOutput := fmt.Sprintf(
				"%s %s %s %s %s %s\n",
				"-t", target,
				"destroy-pipeline",
				"-n",
				"-p", pipelineName,
			)

			Expect(string(output)).To(Equal(expectedOutput))
		})
	})
})
