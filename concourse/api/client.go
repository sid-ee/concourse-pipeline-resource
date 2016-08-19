package api

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/concourse/atc"
	gc "github.com/concourse/go-concourse/concourse"
)

// Enables mocking out of the go-concourse client during tests.
var NewGCClientFunc func(target string, httpClient *http.Client) gc.Client = gc.NewClient

//go:generate counterfeiter . Client

type Client interface {
	Pipelines() ([]Pipeline, error)
	PipelineConfig(pipelineName string) (config atc.Config, rawConfig string, version string, err error)
	SetPipelineConfig(pipelineName string, configVersion string, passedConfig atc.Config) error
	DeletePipeline(pipelineName string) error
}

type client struct {
	gcClient gc.Client
	target   string
}

func NewClient(target string, httpClient *http.Client) Client {
	gcClient := NewGCClientFunc(target, httpClient)

	return &client{gcClient: gcClient, target: target}
}

func HTTPClient(username string, password string, insecure bool) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &http.Client{
		Transport: basicAuthTransport{
			username: username,
			password: password,
			base:     transport,
		},
	}
}

type basicAuthTransport struct {
	username string
	password string

	base http.RoundTripper
}

func (t basicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(t.username, t.password)
	return t.base.RoundTrip(r)
}

func (c client) wrapErr(err error) error {
	return fmt.Errorf(
		"error from target: %s - %s",
		c.target,
		err.Error(),
	)
}
