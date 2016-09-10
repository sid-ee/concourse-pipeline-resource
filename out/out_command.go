package out

import (
	"path/filepath"

	"github.com/concourse/atc"
	"github.com/concourse/fly/template"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
)

const (
	apiPrefix = "/api/v1"
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

//go:generate counterfeiter . PipelineSetter
type PipelineSetter interface {
	SetPipeline(
		teamName string,
		pipelineName string,
		configPath string,
		templateVariables template.Variables,
		templateVariablesFiles []string,
	) error
}

type OutCommand struct {
	logger         Logger
	binaryVersion  string
	apiClient      Client
	sourcesDir     string
	pipelineSetter PipelineSetter
}

func NewOutCommand(
	binaryVersion string,
	logger Logger,
	pipelineSetter PipelineSetter,
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
			p.TeamName,
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

	teamName := input.Source.Teams[0].Name
	apiPipelines, err := c.apiClient.Pipelines(teamName)
	if err != nil {
		return concourse.OutResponse{}, err
	}

	c.logger.Debugf("Found pipelines: %+v\n", pipelines)

	pipelineVersions := make(map[string]string, len(pipelines))

	for _, pipeline := range apiPipelines {
		c.logger.Debugf("Getting pipeline: %s\n", pipeline.Name)
		_, _, version, err := c.apiClient.PipelineConfig(teamName, pipeline.Name)

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
