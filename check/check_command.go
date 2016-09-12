package check

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/concourse/atc"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
)

//go:generate counterfeiter . Client
type Client interface {
	Pipelines(teamName string) ([]api.Pipeline, error)
	PipelineConfig(teamName string, pipelineName string) (config atc.Config, rawConfig string, version string, err error)
}

//go:generate counterfeiter . Logger
type Logger interface {
	Debugf(format string, a ...interface{}) (n int, err error)
}

type CheckCommand struct {
	logger        Logger
	logFilePath   string
	binaryVersion string
	apiClient     Client
}

func NewCheckCommand(
	binaryVersion string,
	logger Logger,
	logFilePath string,
	apiClient Client,
) *CheckCommand {
	return &CheckCommand{
		logger:        logger,
		logFilePath:   logFilePath,
		binaryVersion: binaryVersion,
		apiClient:     apiClient,
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

	c.logger.Debugf("Received input: %+v\n", input)

	c.logger.Debugf("Getting pipelines\n")

	teamPipelines := make(map[string][]api.Pipeline)
	for _, team := range input.Source.Teams {
		pipelines, err := c.apiClient.Pipelines(team.Name)
		if err != nil {
			return nil, err
		}
		teamPipelines[team.Name] = pipelines
	}

	c.logger.Debugf("Found pipelines: %+v\n", teamPipelines)

	pipelineVersions := make(map[string]string)

	for teamName, pipelines := range teamPipelines {
		for _, pipeline := range pipelines {
			c.logger.Debugf("Getting pipeline: %s\n", pipeline.Name)
			_, _, version, err := c.apiClient.PipelineConfig(teamName, pipeline.Name)

			if err != nil {
				return nil, err
			}

			pipelineVersionKey := fmt.Sprintf("%s/%s", teamName, pipeline.Name)

			pipelineVersions[pipelineVersionKey] = version
		}
	}

	out := concourse.CheckResponse{
		pipelineVersions,
	}

	c.logger.Debugf("Returning output: %+v\n", out)

	return out, nil
}
