package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	apiPrefix = "/api/v1"
)

//go:generate counterfeiter . Client

type Client interface {
	Pipelines() ([]Pipeline, error)
}

type client struct {
	target string
}

func NewClient(target string) Client {
	return &client{
		target: target,
	}
}

func (c client) Pipelines() ([]Pipeline, error) {
	targetUrl := fmt.Sprintf(
		"%s%s/pipelines",
		c.target,
		apiPrefix,
	)
	resp, err := http.Get(targetUrl)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected response from %s - status code: %d, expected: %d",
			targetUrl,
			resp.StatusCode,
			http.StatusOK,
		)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// Untested as it is too hard to force ReadAll to return an error
		return nil, err
	}

	var pipelines []Pipeline
	err = json.Unmarshal(b, &pipelines)
	if err != nil {
		return nil, err
	}

	return pipelines, nil
}
