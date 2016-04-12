package acceptance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
)

const (
	outTimeout = 60 * time.Second
)

var _ = Describe("Out", func() {
	var (
		command       *exec.Cmd
		outRequest    concourse.OutRequest
		stdinContents []byte
		sourcesDir    string

		pipelineName           string
		pipelineConfig         string
		pipelineConfigFilename string
		pipelineConfigFilepath string

		varsFileContents string
		varsFileFilename string
		varsFileFilepath string
	)

	BeforeEach(func() {
		var err error
		By("Creating temp directory")
		sourcesDir, err = ioutil.TempDir("", "concourse-pipeline-resource")
		Expect(err).NotTo(HaveOccurred())

		By("Creating random pipeline name")
		pipelineName = fmt.Sprintf("cp-resource-test-%d", time.Now().UnixNano())

		By("Writing pipeline config file")
		pipelineConfig = `---
resources:
- name: concourse-pipeline-resource-repo
  type: git
  uri: https://github.com/robdimsdale/concourse-pipeline-resource.git
  branch: {{foo}}
jobs:
- name: get-concourse-pipeline-resource-repo
  plan:
  - get: concourse-pipeline-resource-repo
`

		pipelineConfigFilename = fmt.Sprintf("%s.yml", pipelineName)
		pipelineConfigFilepath = filepath.Join(sourcesDir, pipelineConfigFilename)
		err = ioutil.WriteFile(pipelineConfigFilepath, []byte(pipelineConfig), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		By("Writing vars file")
		varsFileContents = "foo: bar"

		varsFileFilename = fmt.Sprintf("%s_vars.yml", pipelineName)
		varsFileFilepath = filepath.Join(sourcesDir, varsFileFilename)
		err = ioutil.WriteFile(varsFileFilepath, []byte(varsFileContents), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		By("Creating command object")
		command = exec.Command(outPath, sourcesDir)

		By("Creating default request")
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				Target:   target,
				Username: username,
				Password: password,
			},
			Params: concourse.OutParams{
				Pipelines: []concourse.Pipeline{
					{
						Name:       pipelineName,
						ConfigFile: pipelineConfigFilename,
						VarsFiles: []string{
							varsFileFilename,
						},
					},
				},
			},
		}

		stdinContents, err = json.Marshal(outRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		response, errResponse, err := flyConn.DestroyPipeline(pipelineName)
		if err != nil {
			fmt.Fprintf(GinkgoWriter, "destroy-pipeline failed", string(response), string(errResponse))
		}
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates pipeline, returns valid json", func() {
		By("Running the command")
		session := run(command, stdinContents)
		Eventually(session, outTimeout).Should(gexec.Exit(0))

		By("Outputting a valid json response")
		response := concourse.OutResponse{}
		err := json.Unmarshal(session.Out.Contents(), &response)
		Expect(err).ShouldNot(HaveOccurred())

		By("Validating output contains checksum")
		Expect(response.Version.PipelinesChecksum).NotTo(BeEmpty())
	})

	Context("when validation fails", func() {
		BeforeEach(func() {
			outRequest.Source.Username = ""

			var err error
			stdinContents, err = json.Marshal(outRequest)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("exits with error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited with error")
			Eventually(session, outTimeout).Should(gexec.Exit(1))
			Expect(session.Err).Should(gbytes.Say(".*username.*provided"))
		})
	})
})
