package api

import (
	"fmt"
	"net/http"

	"github.com/concourse/atc"
	gc "github.com/concourse/go-concourse/concourse"
)

func DefaultNewGCClientFunc(url string, teamName string, httpClient *http.Client) ConcourseClient {
	return gc.NewClient(url, httpClient).Team(teamName)
}

// Enables mocking out of the go-concourse client during tests.
var NewGCClientFunc func(url string, teamName string, httpClient *http.Client) ConcourseClient = DefaultNewGCClientFunc

//go:generate counterfeiter . ConcourseClient
type ConcourseClient interface {
	DeletePipeline(pipelineName string) (bool, error)
	ListPipelines() ([]atc.Pipeline, error)
	PipelineConfig(pipelineName string) (atc.Config, atc.RawConfig, string, bool, error)
	CreateOrUpdatePipelineConfig(pipelineName string, configVersion string, passedConfig atc.Config) (bool, bool, []gc.ConfigWarning, error)
}

type Client struct {
	gcClient ConcourseClient
	target   string
}

func NewClient(url string, teamName string, httpClient *http.Client) *Client {
	gcClient := NewGCClientFunc(url, teamName, httpClient)

	return &Client{gcClient: gcClient, target: url}
}

func (c Client) wrapErr(err error) error {
	return fmt.Errorf(
		"error from target: %s - %s",
		c.target,
		err.Error(),
	)
}
