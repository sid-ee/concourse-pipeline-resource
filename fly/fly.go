package fly

import (
	"bytes"
	"fmt"
	"os/exec"

	"crypto/tls"
	"net/http"

	"github.com/robdimsdale/concourse-pipeline-resource/logger"
)

//go:generate counterfeiter . FlyConn

type FlyConn interface {
	Login(url string, username string, password string, insecure bool) ([]byte, error)
	SetPipeline(pipelineName string, configFilepath string, varsFilepaths []string) ([]byte, error)
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
	insecure bool,
) ([]byte, error) {
	extraArgs := ""

	if insecure {
		extraArgs = "-k"
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		http.DefaultClient.Transport = tr
	}
	return f.run(
		"login",
		"-c", url,
		"-u", username,
		"-p", password,
		extraArgs,
	)
}

func (f flyConn) run(args ...string) ([]byte, error) {
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
		// If the command was never started, there will be nothing in the buffers
		return nil, err
	}

	f.logger.Debugf("Waiting for fly command: %v\n", allArgs)
	err = cmd.Wait()
	if err != nil {
		if len(errbuf.Bytes()) > 0 {
			err = fmt.Errorf("%v - %s", err, string(errbuf.Bytes()))
		}
		return outbuf.Bytes(), err
	}

	return outbuf.Bytes(), nil
}

func (f flyConn) SetPipeline(
	pipelineName string,
	configFilepath string,
	varsFilepaths []string,
) ([]byte, error) {
	allArgs := []string{
		"set-pipeline",
		"-n",
		"-p", pipelineName,
		"-c", configFilepath,
	}

	for _, vf := range varsFilepaths {
		allArgs = append(allArgs, "-l", vf)
	}

	return f.run(allArgs...)
}
