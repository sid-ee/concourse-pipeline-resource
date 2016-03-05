package fly

import (
	"os/exec"

	"github.com/robdimsdale/concourse-pipeline-resource/logger"
)

type FlyConn interface {
	Run(...string) ([]byte, error)
}

type flyConn struct {
	target        string
	logger        logger.Logger
	flyBinaryPath string
}

func NewFlyConn(target string, logger logger.Logger, flyBinaryPath string) FlyConn {
	return &flyConn{
		target:        target,
		logger:        logger,
		flyBinaryPath: flyBinaryPath,
	}
}

func (f flyConn) Run(args ...string) ([]byte, error) {
	defaultArgs := []string{
		"-t", f.target,
	}
	allArgs := append(defaultArgs, args...)
	cmd := exec.Command(f.flyBinaryPath, allArgs...)

	f.logger.Debugf("Running fly command: %v\n", allArgs)
	return cmd.CombinedOutput()
}
