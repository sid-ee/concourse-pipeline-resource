package out

import (
	"crypto/md5"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/fly"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/pipelinerunner"
)

const (
	apiPrefix = "/api/v1"
)

type OutCommand struct {
	logger        logger.Logger
	binaryVersion string
	flyConn       fly.FlyConn
	apiClient     api.Client
	sourcesDir    string
}

func NewOutCommand(
	binaryVersion string,
	logger logger.Logger,
	flyConn fly.FlyConn,
	apiClient api.Client,
	sourcesDir string,
) *OutCommand {
	return &OutCommand{
		logger:        logger,
		binaryVersion: binaryVersion,
		flyConn:       flyConn,
		apiClient:     apiClient,
		sourcesDir:    sourcesDir,
	}
}

func (c *OutCommand) Run(input concourse.OutRequest) (concourse.OutResponse, error) {
	if input.Source.Target == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "target")
	}

	if input.Source.Username == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "username")
	}

	if input.Source.Password == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "password")
	}

	if input.Params.Pipelines == nil || len(input.Params.Pipelines) == 0 {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "pipelines")
	}

	for i, p := range input.Params.Pipelines {
		if p.Name == "" {
			return concourse.OutResponse{}, fmt.Errorf("%s must be provided for pipeline[%d]", "name", i)
		}

		if p.ConfigFile == "" {
			return concourse.OutResponse{}, fmt.Errorf("%s must be provided for pipeline[%d]", "config_file", i)
		}
	}

	c.logger.Debugf("Received input: %+v\n", input)

	loginOutput, err := c.flyConn.Login(
		input.Source.Target,
		input.Source.Username,
		input.Source.Password,
	)
	if err != nil {
		c.logger.Debugf("%s\n", string(loginOutput))
		return concourse.OutResponse{}, err
	}

	for _, p := range input.Params.Pipelines {
		configFilepath := filepath.Join(c.sourcesDir, p.ConfigFile)

		runArgs := []string{"set-pipeline", "-n", "-p", p.Name, "-c", configFilepath}
		for _, v := range p.VarsFiles {
			varFilepath := filepath.Join(c.sourcesDir, v)
			runArgs = append(runArgs, "-l", varFilepath)
		}

		_, err := c.flyConn.Run(runArgs...)
		if err != nil {
			return concourse.OutResponse{}, err
		}
	}

	pipelines, err := c.apiClient.Pipelines()
	if err != nil {
		return concourse.OutResponse{}, err
	}

	c.logger.Debugf("Found pipelines: %+v\n", pipelines)

	gpFunc := func(index int, pipeline api.Pipeline) (string, error) {
		b, err := c.flyConn.Run("get-pipeline", "-p", pipeline.Name)
		return string(b), err
	}

	pipelinesContents, err := pipelinerunner.RunForAllPipelines(gpFunc, pipelines, c.logger)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	allContent := strings.Join(pipelinesContents, "")

	pipelinesChecksumString := fmt.Sprintf(
		"%x",
		md5.Sum([]byte(allContent)),
	)
	c.logger.Debugf("pipeline content checksum: %s\n", pipelinesChecksumString)

	metadata := []concourse.Metadata{}

	response := concourse.OutResponse{
		Version: concourse.Version{
			PipelinesChecksum: pipelinesChecksumString,
		},
		Metadata: metadata,
	}

	return response, nil
}
