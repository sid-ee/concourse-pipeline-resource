package fly

import (
	"bytes"
	"fmt"
	"os/exec"

	"crypto/tls"
	"net/http"

	"github.com/concourse/concourse-pipeline-resource/logger"
)

//go:generate counterfeiter . FlyConn

type FlyConn interface {
	Login(url string, teamName string, username string, password string, insecure bool) ([]byte, error)
	GetPipeline(pipelineName string) ([]byte, error)
	SetPipeline(pipelineName string, configFilepath string, varsFilepaths []string) ([]byte, error)
	DestroyPipeline(pipelineName string) ([]byte, error)
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
	teamName string,
	username string,
	password string,
	insecure bool,
) ([]byte, error) {
	args := []string{
		"login",
		"-c", url,
		"-n", teamName,
	}

	if username != "" && password != "" {
		args = append(args, "-u", username, "-p", password)
	}

	if insecure {
		args = append(args, "-k")
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		http.DefaultClient.Transport = tr
	}

	loginOut, err := f.run(args...)
	if err != nil {
		return nil, err
	}

	syncOut, err := f.run("sync")

	if err != nil {
		return nil, err
	}

	return append(loginOut, syncOut...), nil
}

func (f flyConn) run(args ...string) ([]byte, error) {
	if f.target == "" {
		return nil, fmt.Errorf("target cannot be empty in flyConn.run")
	}

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

func (f flyConn) GetPipeline(pipelineName string) ([]byte, error) {
	return f.run(
		"get-pipeline",
		"-p", pipelineName,
	)
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

func (f flyConn) DestroyPipeline(pipelineName string) ([]byte, error) {
	return f.run(
		"destroy-pipeline",
		"-n",
		"-p", pipelineName,
	)
}
