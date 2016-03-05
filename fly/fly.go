package fly

import (
	"os/exec"

	"github.com/robdimsdale/concourse-pipeline-resource/logger"
)

type FlyConn interface {
	Run(...string) ([]byte, error)
}

type flyConn struct {
	target string
	logger logger.Logger
}

func NewFlyConn(target string, logger logger.Logger) FlyConn {
	return &flyConn{
		target: target,
		logger: logger,
	}
}

func (f flyConn) Run(args ...string) ([]byte, error) {
	defaultArgs := []string{
		"-t", f.target,
	}
	allArgs := append(defaultArgs, args...)
	cmd := exec.Command("fly", allArgs...)

	f.logger.Debugf("Running fly command: %v\n", allArgs)
	return cmd.CombinedOutput()
}
