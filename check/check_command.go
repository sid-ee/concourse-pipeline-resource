package check

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/fly"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
)

type CheckCommand struct {
	logger        logger.Logger
	logFilePath   string
	binaryVersion string
	flyConn       fly.FlyConn
}

func NewCheckCommand(
	binaryVersion string,
	logger logger.Logger,
	logFilePath string,
	flyConn fly.FlyConn,
) *CheckCommand {
	return &CheckCommand{
		logger:        logger,
		logFilePath:   logFilePath,
		binaryVersion: binaryVersion,
		flyConn:       flyConn,
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

	loginOutput, err := c.flyConn.Login(
		input.Source.Target,
		input.Source.Username,
		input.Source.Password,
	)
	if err != nil {
		c.logger.Debugf("%s\n", string(loginOutput))
		return nil, err
	}

	apiClient := api.NewClient(input.Source.Target)
	pipelines, err := apiClient.Pipelines()
	if err != nil {
		return nil, err
	}

	c.logger.Debugf("Found pipelines: %+v\n", pipelines)

	var wg sync.WaitGroup
	wg.Add(len(pipelines))

	errChan := make(chan error, len(pipelines))

	pipelinesContents := make([][]byte, len(pipelines))
	for i, p := range pipelines {
		go func(i int, p concourse.Pipeline) {
			defer wg.Done()

			contents, err := c.flyConn.Run("gp", "-p", p.Name)
			if err != nil {
				errChan <- err
			}

			pipelinesContents[i] = contents
		}(i, p)
	}

	c.logger.Debugf("Waiting for all pipelines\n")
	wg.Wait()
	c.logger.Debugf("Waiting for all pipelines complete\n")

	close(errChan)
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	var allPipelineContents []byte
	for i, _ := range pipelinesContents {
		allPipelineContents = append(allPipelineContents, pipelinesContents[i]...)
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
		concourse.Version{
			PipelinesChecksum: pipelinesChecksumString,
		},
	}

	c.logger.Debugf("Returning output: %+v\n", out)

	return out, nil
}
