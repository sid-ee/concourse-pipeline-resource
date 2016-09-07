package api

import (
	"fmt"
	"net/http"

	"github.com/concourse/atc"
	gc "github.com/concourse/go-concourse/concourse"
)

func DefaultNewGCClientFunc(url string, httpClient *http.Client) ConcourseClient {
	return gc.NewClient(url, httpClient).Team("main")
}

// Enables mocking out of the go-concourse client during tests.
var NewGCClientFunc func(url string, httpClient *http.Client) ConcourseClient = DefaultNewGCClientFunc

//go:generate counterfeiter . Client
type Client interface {
	Pipelines() ([]Pipeline, error)
	PipelineConfig(pipelineName string) (config atc.Config, rawConfig string, version string, err error)
	SetPipelineConfig(pipelineName string, configVersion string, passedConfig atc.Config) error
	DeletePipeline(pipelineName string) error
}

//go:generate counterfeiter . ConcourseClient
type ConcourseClient interface {
	DeletePipeline(pipelineName string) (bool, error)
	ListPipelines() ([]atc.Pipeline, error)
	PipelineConfig(pipelineName string) (atc.Config, atc.RawConfig, string, bool, error)
	CreateOrUpdatePipelineConfig(pipelineName string, configVersion string, passedConfig atc.Config) (bool, bool, []gc.ConfigWarning, error)
}

type client struct {
	gcClient ConcourseClient
	target   string
}

func NewClient(url string, httpClient *http.Client) Client {
	gcClient := NewGCClientFunc(url, httpClient)

	return &client{gcClient: gcClient, target: url}
}

func (c client) wrapErr(err error) error {
	return fmt.Errorf(
		"error from target: %s - %s",
		c.target,
		err.Error(),
	)
}
