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
	outTimeout = 5 * time.Second
)

var _ = Describe("Out", func() {
	var (
		command       *exec.Cmd
		outRequest    concourse.OutRequest
		stdinContents []byte
	)

	BeforeEach(func() {
		By("Creating command object")
		command = exec.Command(outPath)

		By("Creating default request")
		outRequest = concourse.OutRequest{
			Source: concourse.Source{
				Target:   target,
				Username: username,
				Password: password,
			},
		}

		var err error
		stdinContents, err = json.Marshal(outRequest)
		Expect(err).ShouldNot(HaveOccurred())
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
