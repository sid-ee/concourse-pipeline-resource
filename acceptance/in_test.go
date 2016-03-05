package acceptance

import (
	"encoding/json"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
)

const (
	inTimeout = 5 * time.Second
)

var _ = Describe("In", func() {
	var (
		command       *exec.Cmd
		inRequest     concourse.InRequest
		stdinContents []byte
	)

	BeforeEach(func() {
		By("Creating command object")
		command = exec.Command(inPath)

		By("Creating default request")
		inRequest = concourse.InRequest{
			Source: concourse.Source{
				Target:   target,
				Username: username,
				Password: password,
			},
			Version: concourse.Version{
				PipelinesChecksum: "",
			},
		}

		var err error
		stdinContents, err = json.Marshal(inRequest)
		Expect(err).ShouldNot(HaveOccurred())
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
