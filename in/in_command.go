package in

import (
	"fmt"

	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
)

const (
	apiPrefix = "/api/v1"
)

type InCommand struct {
	logger        logger.Logger
	binaryVersion string
	flyBinaryPath string
}

func NewInCommand(
	binaryVersion string,
	logger logger.Logger,
	flyBinaryPath string,
) *InCommand {
	return &InCommand{
		logger:        logger,
		binaryVersion: binaryVersion,
		flyBinaryPath: flyBinaryPath,
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

	return concourse.InResponse{}, fmt.Errorf("in is not implemented yet")
}
