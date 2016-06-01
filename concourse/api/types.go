package api

import "github.com/concourse/atc"

type Pipeline struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func pipelineFromATCPipeline(p atc.Pipeline) Pipeline {
	return Pipeline{
		Name: p.Name,
		URL:  p.URL,
	}
}

func pipelinesFromATCPipelines(atcPipelines []atc.Pipeline) []Pipeline {
	pipelines := make([]Pipeline, len(atcPipelines))
	for i, p := range atcPipelines {
		pipelines[i] = pipelineFromATCPipeline(p)
	}

	return pipelines
}
