package acceptance

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/robdimsdale/concourse-pipeline-resource/sanitizer"

	"testing"
)

var (
	inPath    string
	checkPath string
	outPath   string

	target   string
	username string
	password string
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}

var _ = BeforeSuite(func() {
	var err error

	By("Getting target from environment variables")
	target = os.Getenv("TARGET")
	Expect(target).NotTo(BeEmpty(), "$TARGET must be provided")

	By("Getting username from environment variables")
	username = os.Getenv("USERNAME")
	Expect(username).NotTo(BeEmpty(), "$USERNAME must be provided")

	By("Getting password from environment variables")
	password = os.Getenv("PASSWORD")
	Expect(password).NotTo(BeEmpty(), "$PASSWORD must be provided")

	By("Compiling check binary")
	checkPath, err = gexec.Build("github.com/robdimsdale/concourse-pipeline-resource/cmd/check", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Compiling out binary")
	outPath, err = gexec.Build("github.com/robdimsdale/concourse-pipeline-resource/cmd/out", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Compiling in binary")
	inPath, err = gexec.Build("github.com/robdimsdale/concourse-pipeline-resource/cmd/in", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Copying fly to compilation location")
	originalFlyPathPath := os.Getenv("FLY_LOCATION")
	Expect(originalFlyPathPath).NotTo(BeEmpty(), "$FLY_LOCATION must be provided")
	_, err = os.Stat(originalFlyPathPath)
	Expect(err).NotTo(HaveOccurred())

	checkFlyPath := filepath.Join(path.Dir(checkPath), "fly")
	copyFileContents(originalFlyPathPath, checkFlyPath)
	Expect(err).NotTo(HaveOccurred())

	inFlyPath := filepath.Join(path.Dir(inPath), "fly")
	copyFileContents(originalFlyPathPath, inFlyPath)
	Expect(err).NotTo(HaveOccurred())

	outFlyPath := filepath.Join(path.Dir(outPath), "fly")
	copyFileContents(originalFlyPathPath, outFlyPath)
	Expect(err).NotTo(HaveOccurred())

	By("Ensuring copies of fly is executable")
	err = os.Chmod(checkFlyPath, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	err = os.Chmod(inFlyPath, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	err = os.Chmod(outFlyPath, os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	By("Sanitizing acceptance test output")
	sanitized := map[string]string{
		password: "***sanitized-password***",
	}
	sanitizer := sanitizer.NewSanitizer(sanitized, GinkgoWriter)
	GinkgoWriter = sanitizer
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
// See http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func run(command *exec.Cmd, stdinContents []byte) *gexec.Session {
	fmt.Fprintf(GinkgoWriter, "input: %s\n", stdinContents)

	stdin, err := command.StdinPipe()
	Expect(err).ShouldNot(HaveOccurred())

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	_, err = io.WriteString(stdin, string(stdinContents))
	Expect(err).ShouldNot(HaveOccurred())

	err = stdin.Close()
	Expect(err).ShouldNot(HaveOccurred())

	return session
}
