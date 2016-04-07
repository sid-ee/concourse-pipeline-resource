package out

import (
	"fmt"

	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
)

const (
	apiPrefix = "/api/v1"
)

type OutCommand struct {
	logger        logger.Logger
	binaryVersion string
	flyBinaryPath string
}

func NewOutCommand(
	binaryVersion string,
	logger logger.Logger,
	flyBinaryPath string,
) *OutCommand {
	return &OutCommand{
		logger:        logger,
		binaryVersion: binaryVersion,
		flyBinaryPath: flyBinaryPath,
	}
}

func (c *OutCommand) Run(input concourse.OutRequest) (concourse.OutResponse, error) {
	if input.Source.Target == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "target")
	}

	if input.Source.Username == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "username")
	}

	if input.Source.Password == "" {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "password")
	}

	if input.Params.Pipelines == nil || len(input.Params.Pipelines) == 0 {
		return concourse.OutResponse{}, fmt.Errorf("%s must be provided", "pipelines")
	}

	c.logger.Debugf("Received input: %+v\n", input)

	return concourse.OutResponse{}, fmt.Errorf("out is not implemented yet")
}
