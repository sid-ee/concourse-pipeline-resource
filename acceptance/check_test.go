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
	checkTimeout = 5 * time.Second
)

var _ = Describe("Check", func() {
	var (
		command       *exec.Cmd
		checkRequest  concourse.CheckRequest
		stdinContents []byte
	)

	BeforeEach(func() {
		By("Creating command object")
		command = exec.Command(checkPath)

		By("Creating default request")
		checkRequest = concourse.CheckRequest{
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
		stdinContents, err = json.Marshal(checkRequest)
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("when validation fails", func() {
		BeforeEach(func() {
			checkRequest.Source.Username = ""

			var err error
			stdinContents, err = json.Marshal(checkRequest)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("exits with error", func() {
			By("Running the command")
			session := run(command, stdinContents)

			By("Validating command exited with error")
			Eventually(session, checkTimeout).Should(gexec.Exit(1))
			Expect(session.Err).Should(gbytes.Say(".*username.*provided"))
		})
	})
})
