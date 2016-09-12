package in

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

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

type InCommand struct {
	logger        Logger
	binaryVersion string
	apiClient     Client
	downloadDir   string
}

func NewInCommand(
	binaryVersion string,
	logger Logger,
	apiClient Client,
	downloadDir string,
) *InCommand {
	return &InCommand{
		logger:        logger,
		binaryVersion: binaryVersion,
		apiClient:     apiClient,
		downloadDir:   downloadDir,
	}
}

func (c *InCommand) Run(input concourse.InRequest) (concourse.InResponse, error) {
	c.logger.Debugf("Received input: %+v\n", input)

	c.logger.Debugf("Creating download directory: %s\n", c.downloadDir)
	err := os.MkdirAll(c.downloadDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create download directory: %s\n", err.Error())
	}

	c.logger.Debugf("Getting pipelines\n")

	teamPipelines := make(map[string][]api.Pipeline)
	totalPipelineCount := 0
	for _, team := range input.Source.Teams {
		pipelines, err := c.apiClient.Pipelines(team.Name)
		if err != nil {
			return concourse.InResponse{}, err
		}
		teamPipelines[team.Name] = pipelines
		totalPipelineCount += len(pipelines)
	}

	c.logger.Debugf("Found pipelines: %+v\n", teamPipelines)

	var wg sync.WaitGroup
	wg.Add(totalPipelineCount)

	errChan := make(chan error, totalPipelineCount)
	pipelinesWithContents := make([]pipelineWithContent, totalPipelineCount)

	i := 0
	for teamName, pipelines := range teamPipelines {
		for _, p := range pipelines {
			go func(i int, teamName string, p api.Pipeline) {
				defer wg.Done()

				_, config, _, err := c.apiClient.PipelineConfig(teamName, p.Name)
				if err != nil {
					errChan <- err
				}
				pipelinesWithContents[i] = pipelineWithContent{
					name:     p.Name,
					teamName: teamName,
					contents: config,
				}
			}(i, teamName, p)

			i++
		}
	}

	c.logger.Debugf("Waiting for all pipelines\n")
	wg.Wait()
	c.logger.Debugf("Waiting for all pipelines complete\n")

	close(errChan)
	for err := range errChan {
		if err != nil {
			return concourse.InResponse{}, err
		}
	}

	for _, p := range pipelinesWithContents {
		pipelineContentsFilepath := filepath.Join(
			c.downloadDir,
			fmt.Sprintf("%s-%s.yml", p.teamName, p.name),
		)
		c.logger.Debugf("Writing pipeline contents to: %s\n", pipelineContentsFilepath)
		err = ioutil.WriteFile(pipelineContentsFilepath, []byte(p.contents), os.ModePerm)
		if err != nil {
			// Untested as it is too hard to force ioutil.WriteFile to error
			return concourse.InResponse{}, err
		}
	}

	response := concourse.InResponse{
		Version:  input.Version,
		Metadata: []concourse.Metadata{},
	}

	return response, nil
}

type pipelineWithContent struct {
	teamName string
	name     string
	contents string
}
