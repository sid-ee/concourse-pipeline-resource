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
	username string
	password string
}

func NewClient(target string, username string, password string) Client {
	return &client{target: target, username: username, password: password}
}

func (c client) Pipelines() ([]Pipeline, error) {
	client := &http.Client{}
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s%s/pipelines",
			c.target,
			apiPrefix,
		),
		nil)

	req.SetBasicAuth(c.username, c.password)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected response status code: %d, expected: %d",
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
