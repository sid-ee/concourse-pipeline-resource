package pipelinerunner

import (
	"sync"

	"github.com/concourse/concourse-pipeline-resource/concourse/api"
	"github.com/concourse/concourse-pipeline-resource/logger"
)

func RunForAllPipelines(
	function func(index int, pipeline api.Pipeline) (string, error),
	pipelines []api.Pipeline,
	logger logger.Logger,
) ([]string, error) {
	var wg sync.WaitGroup
	wg.Add(len(pipelines))

	errChan := make(chan error, len(pipelines))

	pipelinesContents := make([]string, len(pipelines))
	for i, p := range pipelines {
		go func(i int, p api.Pipeline) {
			defer wg.Done()

			contents, err := function(i, p)
			if err != nil {
				errChan <- err
			}

			pipelinesContents[i] = contents
		}(i, p)
	}

	logger.Debugf("Waiting for all pipelines\n")
	wg.Wait()
	logger.Debugf("Waiting for all pipelines complete\n")

	close(errChan)
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return pipelinesContents, nil
}
