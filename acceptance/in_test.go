package acceptance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/concourse/concourse-pipeline-resource/concourse"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const (
	inTimeout = 40 * time.Second
)

var _ = Describe("In", func() {
	var (
		command       *exec.Cmd
		inRequest     concourse.InRequest
		stdinContents []byte
		destDirectory string
	)

	BeforeEach(func() {
		var err error

		By("Restoring environment variables")
		RestoreEnvVars()

		By("Creating temp directory")
		destDirectory, err = ioutil.TempDir("", "concourse-pipeline-resource")
		Expect(err).NotTo(HaveOccurred())

		By("Creating command object")
		command = exec.Command(inPath, destDirectory)

		By("Creating default request")

		inRequest = concourse.InRequest{
			Source: concourse.Source{
				Target:   target,
				Insecure: fmt.Sprintf("%t", insecure),
				Teams: []concourse.Team{
					{
						Name:     teamName,
						Username: username,
						Password: password,
					},
				},
			},
			Version: concourse.Version{
				"some-pipeline":       "some-pipeline-version",
				"some-other-pipeline": "some-other-pipeline-version",
			},
		}

		stdinContents, err = json.Marshal(inRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Describe("successful behavior", func() {
		Context("with a test pipeline", func() {
			var (
				testPipelineDir      string
				testPipelineFilePath string
				testPipelineName     string
				testPipelineCreated  bool
			)

			BeforeEach(func() {
				var err error

				By("Creating temp directory")
				testPipelineDir, err = ioutil.TempDir("", "concourse-pipeline-resource")
				Expect(err).NotTo(HaveOccurred())

				By("Creating random pipeline name")
				testPipelineName = fmt.Sprintf("cp-resource-test-%d", time.Now().UnixNano())

				By("Creating test pipeline config file")
				testPipelineFilePath, err = CreateTestPipelineConfigFile(testPipelineDir, testPipelineName)
				Expect(err).NotTo(HaveOccurred())

				By("Creating a test pipeline")
				err = SetTestPipeline(testPipelineName, testPipelineFilePath)
				Expect(err).NotTo(HaveOccurred())
				testPipelineCreated = true
			})

			AfterEach(func() {
				if testPipelineCreated {
					_, err := flyCommand.DestroyPipeline(testPipelineName)
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("downloads all pipeline configs to the target directory", func() {
				By("Running the command")
				session := run(command, stdinContents)

				Eventually(session, inTimeout).Should(gexec.Exit(0))

				files, err := ioutil.ReadDir(destDirectory)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(files)).To(BeNumerically(">", 0))
				for _, file := range files {
					Expect(file.Name()).To(MatchRegexp(".*\\.yml"))
					Expect(file.Size()).To(BeNumerically(">", 0))
				}
			})
		})

		It("returns valid json", func() {
			By("Running the command")
			session := run(command, stdinContents)
			Eventually(session, inTimeout).Should(gexec.Exit(0))

			By("Outputting a valid json response")
			response := concourse.InResponse{}
			err := json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).ShouldNot(HaveOccurred())

			By("Validating output contains pipeline versions")
			Expect(len(response.Version)).To(BeNumerically(">", 0))
			for k, v := range response.Version {
				Expect(k).NotTo(BeEmpty())
				Expect(v).NotTo(BeEmpty())
			}
		})

		Context("target not provided by source", func() {
			BeforeEach(func() {
				var err error
				err = os.Setenv("ATC_EXTERNAL_URL", inRequest.Source.Target)
				Expect(err).ShouldNot(HaveOccurred())

				inRequest.Source.Target = ""

				stdinContents, err = json.Marshal(inRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("uses ATC_EXTERNAL_URL instead", func() {
				By("Running the command")
				session := run(command, stdinContents)

				By("Validating command ran successfully")
				Eventually(session, inTimeout).Should(gexec.Exit(0))
			})
		})

		Context("target not provided at all", func() {
			BeforeEach(func() {
				inRequest.Source.Target = ""

				var err error
				stdinContents, err = json.Marshal(inRequest)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("exits with error", func() {
				By("Running the command")
				session := run(command, stdinContents)

				By("Validating command exited with error")
				Eventually(session, inTimeout).Should(gexec.Exit(1))
				Expect(session.Err).Should(gbytes.Say("target must be provided in source"))
			})
		})
	})

	Context("when validation fails", func() {
		BeforeEach(func() {
			inRequest.Source.Teams = nil

			var err error
			stdinContents, err = json.Marshal(inRequest)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("exits with error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited with error")
			Eventually(session, inTimeout).Should(gexec.Exit(1))
			Expect(session.Err).Should(gbytes.Say(".*teams.*provided"))
		})
	})
})
