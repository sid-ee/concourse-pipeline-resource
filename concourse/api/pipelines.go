package api

import (
	"crypto/tls"
	"fmt"
	"net/http"

	gc "github.com/concourse/go-concourse/concourse"
)

//go:generate counterfeiter . Client

type Client interface {
	Pipelines() ([]Pipeline, error)
}

type client struct {
	gcClient gc.Client
	target   string
}

func NewClient(target string, username string, password string, insecure bool) Client {
	httpClient := &http.Client{}

	if insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		httpClient.Transport = tr
	}

	gcClient := gc.NewClient(target, httpClient)

	return &client{gcClient: gcClient, target: target}
}

func (c client) Pipelines() ([]Pipeline, error) {
	atcPipelines, err := c.gcClient.ListPipelines()
	if err != nil {
		return nil, c.wrapErr(err)
	}

	return pipelinesFromATCPipelines(atcPipelines), nil
}

func (c client) wrapErr(err error) error {
	return fmt.Errorf(
		"error from target: %s - %s",
		c.target,
		err.Error(),
	)
}
