package out

import (
	"path/filepath"

	"github.com/concourse/atc"
	"github.com/concourse/fly/template"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/out/helpers"
)

const (
	apiPrefix = "/api/v1"
)

//go:generate counterfeiter . Client
type Client interface {
	Pipelines() ([]api.Pipeline, error)
	PipelineConfig(pipelineName string) (config atc.Config, rawConfig string, version string, err error)
}

type OutCommand struct {
	logger         logger.Logger
	binaryVersion  string
	apiClient      Client
	sourcesDir     string
	pipelineSetter helpers.PipelineSetter
}

func NewOutCommand(
	binaryVersion string,
	logger logger.Logger,
	pipelineSetter helpers.PipelineSetter,
	apiClient Client,
	sourcesDir string,
) *OutCommand {
	return &OutCommand{
		logger:         logger,
		binaryVersion:  binaryVersion,
		pipelineSetter: pipelineSetter,
		apiClient:      apiClient,
		sourcesDir:     sourcesDir,
	}
}

func (c *OutCommand) Run(input concourse.OutRequest) (concourse.OutResponse, error) {
	c.logger.Debugf("Received input: %+v\n", input)

	pipelines := input.Params.Pipelines

	c.logger.Debugf("Input pipelines: %+v\n", pipelines)

	c.logger.Debugf("Setting pipelines\n")
	for _, p := range pipelines {
		configFilepath := filepath.Join(c.sourcesDir, p.ConfigFile)

		var varsFilepaths []string
		for _, v := range p.VarsFiles {
			varFilepath := filepath.Join(c.sourcesDir, v)
			varsFilepaths = append(varsFilepaths, varFilepath)
		}

		var templateVariables template.Variables
		err := c.pipelineSetter.SetPipeline(
			p.Name,
			configFilepath,
			templateVariables,
			varsFilepaths,
		)
		if err != nil {
			return concourse.OutResponse{}, err
		}
	}
	c.logger.Debugf("Setting pipelines complete\n")

	c.logger.Debugf("Getting pipelines\n")

	apiPipelines, err := c.apiClient.Pipelines()
	if err != nil {
		return concourse.OutResponse{}, err
	}

	c.logger.Debugf("Found pipelines: %+v\n", pipelines)

	pipelineVersions := make(map[string]string, len(pipelines))

	for _, pipeline := range apiPipelines {
		c.logger.Debugf("Getting pipeline: %s\n", pipeline.Name)
		_, _, version, err := c.apiClient.PipelineConfig(pipeline.Name)

		if err != nil {
			return concourse.OutResponse{}, err
		}

		pipelineVersions[pipeline.Name] = version
	}

	response := concourse.OutResponse{
		Version:  pipelineVersions,
		Metadata: []concourse.Metadata{},
	}

	return response, nil
}
