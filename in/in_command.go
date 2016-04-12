package in

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/concourse/api"
	"github.com/robdimsdale/concourse-pipeline-resource/fly"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
)

const (
	apiPrefix = "/api/v1"
)

type InCommand struct {
	logger        logger.Logger
	binaryVersion string
	flyConn       fly.FlyConn
	apiClient     api.Client
	downloadDir   string
}

func NewInCommand(
	binaryVersion string,
	logger logger.Logger,
	flyConn fly.FlyConn,
	apiClient api.Client,
	downloadDir string,
) *InCommand {
	return &InCommand{
		logger:        logger,
		binaryVersion: binaryVersion,
		flyConn:       flyConn,
		apiClient:     apiClient,
		downloadDir:   downloadDir,
	}
}

func (c *InCommand) Run(input concourse.InRequest) (concourse.InResponse, error) {
	c.logger.Debugf("Received input: %+v\n", input)

	loginOutput, loginErr, err := c.flyConn.Login(
		input.Source.Target,
		input.Source.Username,
		input.Source.Password,
	)
	if err != nil {
		c.logger.Debugf("%s\n", string(loginOutput))
		c.logger.Debugf("%s\n", string(loginErr))
		return concourse.InResponse{}, err
	}

	c.logger.Debugf("Creating download directory: %s\n", c.downloadDir)
	err = os.MkdirAll(c.downloadDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create download directory: %s\n", err.Error())
	}

	c.logger.Debugf("Getting pipelines\n")

	pipelines, err := c.apiClient.Pipelines()
	if err != nil {
		return concourse.InResponse{}, err
	}

	c.logger.Debugf("Found pipelines: %+v\n", pipelines)

	var wg sync.WaitGroup
	wg.Add(len(pipelines))

	errChan := make(chan error, len(pipelines))

	pipelinesWithContents := make([]pipelineWithContent, len(pipelines))
	for i, p := range pipelines {
		go func(i int, p api.Pipeline) {
			defer wg.Done()

			outContents, _, err := c.flyConn.GetPipeline(p.Name)
			if err != nil {
				errChan <- err
			}
			pipelinesWithContents[i] = pipelineWithContent{
				name:     p.Name,
				contents: outContents,
			}
		}(i, p)
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
		pipelineContentsFilepath := filepath.Join(c.downloadDir, fmt.Sprintf("%s.yml", p.name))
		c.logger.Debugf("Writing pipeline contents to: %s\n", pipelineContentsFilepath)
		err = ioutil.WriteFile(pipelineContentsFilepath, p.contents, os.ModePerm)
		if err != nil {
			// Untested as it is too hard to force ioutil.WriteFile to error
			return concourse.InResponse{}, err
		}
	}

	metadata := make([]concourse.Metadata, len(pipelinesWithContents))

	response := concourse.InResponse{
		Version: concourse.Version{
			PipelinesChecksum: input.Version.PipelinesChecksum,
		},
		Metadata: metadata,
	}

	return response, nil
}

type pipelineWithContent struct {
	name     string
	contents []byte
}
