package acceptance

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/concourse/concourse-pipeline-resource/fly"
	"github.com/concourse/concourse-pipeline-resource/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/robdimsdale/sanitizer"

	"testing"
)

const (
	teamName = "main"
)

type Client interface {
	DeletePipeline(teamName string, pipelineName string) error
}

var (
	inPath    string
	checkPath string
	outPath   string

	target   string
	username string
	password string
	insecure bool

	env map[string]string

	flyCommand fly.Command
)

func CaptureEnvVars() map[string]string {
	capturedEnv := make(map[string]string)
	// get list of current Key=Value
	currentEnv := os.Environ()
	// iterate and split into "Key": "Value"
	for _, envVarItem := range currentEnv {
		// split into before and after the first =
		envVarItemKeyValue := strings.SplitN(envVarItem, "=", 2)
		envVarItemKey := envVarItemKeyValue[0]
		envVarItemValue := envVarItemKeyValue[1]
		capturedEnv[envVarItemKey] = envVarItemValue
	}
	return capturedEnv
}

func RestoreEnvVars() {
	os.Clearenv()
	var err error
	for envVarKey, envVarValue := range env {
		err = os.Setenv(envVarKey, envVarValue)
		if err != nil {
			fmt.Fprintln(GinkgoWriter, err.Error())
		}
	}
}

func CreateTestPipelineConfigFile(dirPath, pipelineName string) (string, error) {
	var err error

	pipelineConfig := `---
resources:
- name: concourse-pipeline-resource-repo
  type: git
  source:
    uri: https://github.com/concourse/concourse-pipeline-resource.git
    branch: master
jobs:
- name: get-concourse-pipeline-resource-repo
  plan:
  - get: concourse-pipeline-resource-repo
`

	pipelineConfigFileName := fmt.Sprintf("%s.yml", pipelineName)
	pipelineConfigFilePath := filepath.Join(dirPath, pipelineConfigFileName)
	err = ioutil.WriteFile(pipelineConfigFilePath, []byte(pipelineConfig), os.ModePerm)
	return pipelineConfigFilePath, err
}

func SetTestPipeline(pipelineName string, configFilePath string) error {
	var err error
	var setOutput []byte
	setOutput, err = flyCommand.SetPipeline(pipelineName, configFilePath, nil, nil)
	fmt.Fprintf(GinkgoWriter, "pipeline '%s' set; output:\n\n%s\n", pipelineName, string(setOutput))
	return err
}

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}

var _ = BeforeSuite(func() {
	var err error

	By("Capturing current environment variables")
	env = CaptureEnvVars()

	By("Getting target from environment variables")
	target = os.Getenv("TARGET")
	Expect(target).NotTo(BeEmpty(), "$TARGET must be provided")

	By("Getting username from environment variables")
	username = os.Getenv("USERNAME")

	By("Getting password from environment variables")
	password = os.Getenv("PASSWORD")

	insecureFlag := os.Getenv("INSECURE")
	if insecureFlag != "" {
		By("Getting insecure from environment variables")
		insecure, err = strconv.ParseBool(insecureFlag)
		Expect(err).NotTo(HaveOccurred())
	}

	By("Compiling check binary")
	checkPath, err = gexec.Build("github.com/concourse/concourse-pipeline-resource/cmd/check", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Compiling out binary")
	outPath, err = gexec.Build("github.com/concourse/concourse-pipeline-resource/cmd/out", "-race")
	Expect(err).NotTo(HaveOccurred())

	By("Compiling in binary")
	inPath, err = gexec.Build("github.com/concourse/concourse-pipeline-resource/cmd/in", "-race")
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

	By("Creating fly connection")
	l := logger.NewLogger(sanitizer)
	flyCommand = fly.NewCommand("concourse-pipeline-resource-target", l, inFlyPath)

	By("Logging in with fly")
	_, err = flyCommand.Login(target, teamName, username, password, insecure)
	Expect(err).NotTo(HaveOccurred())
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
