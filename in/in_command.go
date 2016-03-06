package in

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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
	downloadDir   string
}

func NewInCommand(
	binaryVersion string,
	logger logger.Logger,
	flyConn fly.FlyConn,
	downloadDir string,
) *InCommand {
	return &InCommand{
		logger:        logger,
		binaryVersion: binaryVersion,
		flyConn:       flyConn,
		downloadDir:   downloadDir,
	}
}

func (c *InCommand) Run(input concourse.InRequest) (concourse.InResponse, error) {
	if input.Source.Target == "" {
		return concourse.InResponse{}, fmt.Errorf("%s must be provided", "target")
	}

	if input.Source.Username == "" {
		return concourse.InResponse{}, fmt.Errorf("%s must be provided", "username")
	}

	if input.Source.Password == "" {
		return concourse.InResponse{}, fmt.Errorf("%s must be provided", "password")
	}

	c.logger.Debugf("Received input: %+v\n", input)

	loginOutput, err := c.flyConn.Login(
		input.Source.Target,
		input.Source.Username,
		input.Source.Password,
	)
	if err != nil {
		c.logger.Debugf("%s\n", string(loginOutput))
		return concourse.InResponse{}, err
	}

	c.logger.Debugf("Creating download directory: %s\n", c.downloadDir)
	err = os.MkdirAll(c.downloadDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create download directory: %s\n", err.Error())
	}

	apiClient := api.NewClient(input.Source.Target)
	pipelines, err := apiClient.Pipelines()
	if err != nil {
		return concourse.InResponse{}, err
	}

	c.logger.Debugf("Found pipelines: %+v\n", pipelines)

	for _, p := range pipelines {
		pipelineContents, err := c.flyConn.Run("gp", "-p", p.Name)
		if err != nil {
			return concourse.InResponse{}, err
		}

		pipelineContentsFilepath := filepath.Join(c.downloadDir, fmt.Sprintf("%s.yml", p.Name))
		c.logger.Debugf("Writing pipeline contents to: %s\n", pipelineContentsFilepath)
		err = ioutil.WriteFile(pipelineContentsFilepath, pipelineContents, os.ModePerm)
		if err != nil {
			// Untested as it is too hard to force ioutil.WriteFile to error
			return concourse.InResponse{}, err
		}
	}

	return concourse.InResponse{}, nil
}
