package check

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/fly"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
)

const (
	apiPrefix = "/api/v1"
)

type CheckCommand struct {
	logger        logger.Logger
	logFilePath   string
	binaryVersion string
	flyBinaryPath string
}

func NewCheckCommand(
	binaryVersion string,
	logger logger.Logger,
	logFilePath string,
	flyBinaryPath string,
) *CheckCommand {
	return &CheckCommand{
		logger:        logger,
		logFilePath:   logFilePath,
		binaryVersion: binaryVersion,
		flyBinaryPath: flyBinaryPath,
	}
}

func (c *CheckCommand) Run(input concourse.CheckRequest) (concourse.CheckResponse, error) {
	logDir := filepath.Dir(c.logFilePath)
	existingLogFiles, err := filepath.Glob(filepath.Join(logDir, "concourse-pipeline-resource-check.log*"))
	if err != nil {
		// This is untested because the only error returned by filepath.Glob is a
		// malformed glob, and this glob is hard-coded to be correct.
		return nil, err
	}

	for _, f := range existingLogFiles {
		if filepath.Base(f) != filepath.Base(c.logFilePath) {
			c.logger.Debugf("Removing existing log file: %s\n", f)
			err := os.Remove(f)
			if err != nil {
				// This is untested because it is too hard to force os.Remove to return
				// an error.
				return nil, err
			}
		}
	}

	if input.Source.Target == "" {
		return nil, fmt.Errorf("%s must be provided", "target")
	}

	if input.Source.Username == "" {
		return nil, fmt.Errorf("%s must be provided", "username")
	}

	if input.Source.Password == "" {
		return nil, fmt.Errorf("%s must be provided", "password")
	}

	c.logger.Debugf("Received input: %+v\n", input)

	flyConn := fly.NewFlyConn("concourse-pipeline-resource-target", c.logger, c.flyBinaryPath)

	loginOutput, err := flyConn.Run(
		"login",
		"-c", input.Source.Target,
		"-u", input.Source.Username,
		"-p", input.Source.Password,
	)
	if err != nil {
		c.logger.Debugf("%s\n", string(loginOutput))
		panic(err)
	}

	resp, err := http.Get(fmt.Sprintf(
		"%s%s/pipelines",
		input.Source.Target,
		apiPrefix,
	))
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var pipelines []concourse.Pipeline
	err = json.Unmarshal(b, &pipelines)
	if err != nil {
		panic(err)
	}

	c.logger.Debugf("Found pipelines: %+v\n", pipelines)

	var allPipelineContents []byte
	for _, p := range pipelines {
		pipelineContents, err := flyConn.Run("gp", "-p", p.Name)
		if err != nil {
			panic(err)
		}

		allPipelineContents = append(allPipelineContents, pipelineContents...)
	}

	pipelinesChecksumString := fmt.Sprintf(
		"%x",
		md5.Sum(allPipelineContents),
	)
	c.logger.Debugf("all pipeline contents:\n%s\n", string(allPipelineContents))

	if input.Version.PipelinesChecksum == pipelinesChecksumString {
		c.logger.Debugf("No new versions found\n")
		return concourse.CheckResponse{}, nil
	}

	out := concourse.CheckResponse{
		{PipelinesChecksum: pipelinesChecksumString},
	}

	c.logger.Debugf("Returning output: %+v\n", out)

	return out, nil
}
