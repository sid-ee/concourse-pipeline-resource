package fly

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/robdimsdale/concourse-pipeline-resource/logger"
)

//go:generate counterfeiter . FlyConn

type FlyConn interface {
	Login(target string, username string, password string) ([]byte, []byte, error)
	GetPipeline(pipelineName string) ([]byte, []byte, error)
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

func (f flyConn) Login(
	url string,
	username string,
	password string,
) ([]byte, []byte, error) {
	return f.run(
		"login",
		"-c", url,
		"-u", username,
		"-p", password,
	)
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

func (f flyConn) run(args ...string) ([]byte, []byte, error) {
	defaultArgs := []string{
		"-t", f.target,
	}
	allArgs := append(defaultArgs, args...)
	cmd := exec.Command(f.flyBinaryPath, allArgs...)

	outbuf := bytes.NewBuffer(nil)
	errbuf := bytes.NewBuffer(nil)
	cmd.Stdout = outbuf
	cmd.Stderr = errbuf

	f.logger.Debugf("Starting fly command: %v\n", allArgs)
	err := cmd.Start()
	if err != nil {
		if len(errbuf.Bytes()) > 0 {
			err = fmt.Errorf("%v - %s", err, string(errbuf.Bytes()))
		}
		return outbuf.Bytes(), errbuf.Bytes(), err
	}

	f.logger.Debugf("Waiting for fly command: %v\n", allArgs)
	err = cmd.Wait()
	if err != nil {
		if len(errbuf.Bytes()) > 0 {
			err = fmt.Errorf("%v - %s", err, string(errbuf.Bytes()))
		}
		return outbuf.Bytes(), errbuf.Bytes(), err
	}

	return outbuf.Bytes(), errbuf.Bytes(), nil
}

func (f flyConn) GetPipeline(pipelineName string) ([]byte, []byte, error) {
	return f.run(
		"get-pipeline",
		"-p", pipelineName,
	)
}
