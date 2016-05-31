package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const (
	apiPrefix = "/api/v1"
)

//go:generate counterfeiter . Client

type Client interface {
	Pipelines() ([]Pipeline, error)
}

type client struct {
	target   string
	username string
	password string
	insecure string
}

func NewClient(target string, username string, password string, insecure string) Client {
	return &client{target: target, username: username, password: password, insecure: insecure}
}

func (c client) Pipelines() ([]Pipeline, error) {
	targetUrl := fmt.Sprintf(
		"%s%s/pipelines",
		c.target,
		apiPrefix,
	)

	insecure := false
	if c.insecure != "" {
		var err error
		insecure, err = strconv.ParseBool(c.insecure)
		if err != nil {
			log.Fatalln("Invalid value for insecure: %v", c.insecure)
		}
	}

	client := &http.Client{}

	if insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		}
		client.Transport = tr
	}

	req, err := http.NewRequest(
		"GET",
		targetUrl,
		nil)

	req.SetBasicAuth(c.username, c.password)
	resp, err := client.Do(req)

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
