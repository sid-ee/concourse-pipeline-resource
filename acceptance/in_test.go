package acceptance

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
)

const (
	inTimeout = 20 * time.Second
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

		By("Creating temp directory")
		destDirectory, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())

		By("Creating command object")
		command = exec.Command(inPath, destDirectory)

		By("Creating default request")
		inRequest = concourse.InRequest{
			Source: concourse.Source{
				Target:   target,
				Username: username,
				Password: password,
			},
			Version: concourse.Version{
				PipelinesChecksum: "some-pipeline-checksum",
			},
		}

		stdinContents, err = json.Marshal(inRequest)
		Expect(err).ShouldNot(HaveOccurred())
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

	It("returns valid json", func() {
		By("Running the command")
		session := run(command, stdinContents)
		Eventually(session, inTimeout).Should(gexec.Exit(0))

		By("Outputting a valid json response")
		response := concourse.InResponse{}
		err := json.Unmarshal(session.Out.Contents(), &response)
		Expect(err).ShouldNot(HaveOccurred())

		By("Validating output contains product version")
		Expect(response.Version.PipelinesChecksum).To(Equal(inRequest.Version.PipelinesChecksum))
	})

	Context("when validation fails", func() {
		BeforeEach(func() {
			inRequest.Source.Username = ""

			var err error
			stdinContents, err = json.Marshal(inRequest)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("exits with error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited with error")
			Eventually(session, inTimeout).Should(gexec.Exit(1))
			Expect(session.Err).Should(gbytes.Say(".*username.*provided"))
		})
	})
})
