package api

import "fmt"

func (c client) Pipelines() ([]Pipeline, error) {
	atcPipelines, err := c.gcClient.ListPipelines()
	if err != nil {
		return nil, c.wrapErr(err)
	}

	return pipelinesFromATCPipelines(atcPipelines), nil
}

func (c client) PipelineConfig(pipelineName string) (string, error) {
	_, atcRawConfig, _, exists, err := c.gcClient.PipelineConfig(pipelineName)
	if err != nil {
		return "", c.wrapErr(err)
	}

	if !exists {
		err := fmt.Errorf("Pipeline not found: %s", pipelineName)
		return "", c.wrapErr(err)
	}

	return atcRawConfig.String(), nil
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
