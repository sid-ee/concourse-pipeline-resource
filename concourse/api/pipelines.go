package api

import (
	"fmt"

	"github.com/concourse/atc"
)

func (c Client) Pipelines(teamName string) ([]Pipeline, error) {
	atcPipelines, err := c.gcClients[teamName].ListPipelines()
	if err != nil {
		return nil, c.wrapErr(err)
	}

	return pipelinesFromATCPipelines(atcPipelines), nil
}

func (c Client) PipelineConfig(teamName string, pipelineName string) (atc.Config, string, string, error) {
	atcConfig, atcRawConfig, configVersion, exists, err :=
		c.gcClients[teamName].PipelineConfig(pipelineName)
	if err != nil {
		return atc.Config{}, "", "", c.wrapErr(err)
	}

	if !exists {
		err := fmt.Errorf("Pipeline not found: %s", pipelineName)
		return atc.Config{}, "", "", c.wrapErr(err)
	}

	return atcConfig, atcRawConfig.String(), configVersion, nil
}

func (c Client) SetPipelineConfig(
	teamName string,
	pipelineName string,
	configVersion string,
	passedConfig atc.Config,
) error {
	created, updated, _, err := c.gcClients[teamName].CreateOrUpdatePipelineConfig(
		pipelineName,
		configVersion,
		passedConfig,
	)
	if err != nil {
		return c.wrapErr(err)
	}

	if !created && !updated {
		err := fmt.Errorf("Pipeline not created or updated: %s", pipelineName)
		return c.wrapErr(err)
	}

	return nil
}

func (c Client) DeletePipeline(teamName string, pipelineName string) error {
	exists, err := c.gcClients[teamName].DeletePipeline(pipelineName)
	if err != nil {
		return c.wrapErr(err)
	}

	if !exists {
		err := fmt.Errorf("Pipeline not found: %s", pipelineName)
		return c.wrapErr(err)
	}

	return nil
}
