package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/concourse/concourse-pipeline-resource/concourse"
)

const (
	apiPrefix = "/api/v1"
)

//go:generate counterfeiter . Client

type Client interface {
	Pipelines(string) ([]Pipeline, error)
}

type client struct {
	target   string
	insecure bool
	teams    []concourse.Team
}

func NewClient(target string, insecure bool, teams []concourse.Team) Client {
	return &client{
		target:   target,
		insecure: insecure,
		teams:    teams,
	}
}

type AuthToken struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (c client) Pipelines(teamName string) ([]Pipeline, error) {
	var team *concourse.Team

	for _, t := range c.teams {
		if t.Name == teamName {
			team = &t
		}
	}

	if team == nil {
		return nil, fmt.Errorf("team not configured: %s", teamName)
	}

	client := &http.Client{}

	if c.insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		client.Transport = tr
	}

	tokenUrl := fmt.Sprintf(
		"%s%s/teams/%s/auth/token",
		c.target,
		apiPrefix,
		teamName,
	)

	req, err := http.NewRequest(
		"GET",
		tokenUrl,
		nil)

	req.SetBasicAuth(team.Username, team.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected response from %s - status code: %d, expected: %d",
			tokenUrl,
			resp.StatusCode,
			http.StatusOK,
		)
	}

	var token AuthToken
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, err
	}

	targetUrl := fmt.Sprintf(
		"%s%s/teams/%s/pipelines",
		c.target,
		apiPrefix,
		teamName,
	)

	req, err = http.NewRequest(
		"GET",
		targetUrl,
		nil)

	req.Header.Set("Authorization", fmt.Sprintf("%s %s", token.Type, token.Value))

	resp, err = client.Do(req)

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
