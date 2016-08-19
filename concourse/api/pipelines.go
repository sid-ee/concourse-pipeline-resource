package api

import (
	"fmt"

	"github.com/concourse/atc"
)

func (c client) Pipelines() ([]Pipeline, error) {
	atcPipelines, err := c.gcClient.ListPipelines()
	if err != nil {
		return nil, c.wrapErr(err)
	}

	return pipelinesFromATCPipelines(atcPipelines), nil
}

func (c client) PipelineConfig(pipelineName string) (atc.Config, string, string, error) {
	atcConfig, atcRawConfig, configVersion, exists, err :=
		c.gcClient.PipelineConfig(pipelineName)
	if err != nil {
		return atc.Config{}, "", "", c.wrapErr(err)
	}

	if !exists {
		err := fmt.Errorf("Pipeline not found: %s", pipelineName)
		return atc.Config{}, "", "", c.wrapErr(err)
	}

	return atcConfig, atcRawConfig.String(), configVersion, nil
}

func (c client) SetPipelineConfig(pipelineName string, configVersion string, passedConfig atc.Config) error {
	created, updated, _, err := c.gcClient.CreateOrUpdatePipelineConfig(
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

func (c client) DeletePipeline(pipelineName string) error {
	exists, err := c.gcClient.DeletePipeline(pipelineName)
	if err != nil {
		return c.wrapErr(err)
	}

	if !exists {
		err := fmt.Errorf("Pipeline not found: %s", pipelineName)
		return c.wrapErr(err)
	}

	return nil
}
