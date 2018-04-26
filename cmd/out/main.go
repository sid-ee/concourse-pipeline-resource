package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/concourse/concourse-pipeline-resource/cmd/out/filereader"
	"github.com/concourse/concourse-pipeline-resource/concourse"
	"github.com/concourse/concourse-pipeline-resource/concourse/api"
	"github.com/concourse/concourse-pipeline-resource/fly"
	"github.com/concourse/concourse-pipeline-resource/logger"
	"github.com/concourse/concourse-pipeline-resource/out"
	"github.com/concourse/concourse-pipeline-resource/validator"
	"github.com/robdimsdale/sanitizer"
)

const (
	flyBinaryName        = "fly"
	atcExternalURLEnvKey = "ATC_EXTERNAL_URL"
)

var (
	l logger.Logger
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln(fmt.Sprintf(
			"not enough args - usage: %s <sources directory>", os.Args[0]))
	}

	sourcesDir := os.Args[1]

	outDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}

	var input concourse.OutRequest

	logFile, err := ioutil.TempFile("", "concourse-pipeline-resource-out.log")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Fprintf(os.Stderr, "Logging to %s\n", logFile.Name())

	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		fmt.Fprintf(logFile, "Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	l = logger.NewLogger(sanitizer)

	flyBinaryPath := filepath.Join(outDir, flyBinaryName)

	if input.Source.Target == "" {
		input.Source.Target = os.Getenv(atcExternalURLEnvKey)
	}

	flyCommand := fly.NewCommand(input.Source.Target, l, flyBinaryPath)

	err = validator.ValidateOut(input)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	if input.Params.PipelinesFile != "" {
		pipelinesFromFile, err := filereader.PipelinesFromFile(input.Params.PipelinesFile, sourcesDir)
		if err != nil {
			l.Debugf("Exiting with error: %v\n", err)
			log.Fatalln(err)
		}

		input.Params.PipelinesFile = ""
		input.Params.Pipelines = pipelinesFromFile
	}

	// Validate contents of pipelines file
	err = validator.ValidateOut(input)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	insecure := false
	if input.Source.Insecure != "" {
		var err error
		insecure, err = strconv.ParseBool(input.Source.Insecure)
		if err != nil {
			log.Fatalf("Invalid value for insecure: %v", input.Source.Insecure)
		}
	}

	apiClient := api.NewClient(input.Source.Target, insecure, input.Source.Teams)
	response, err := out.NewCommand(l, flyCommand, apiClient, sourcesDir).Run(input)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	l.Debugf("Returning output: %+v\n", response)

	err = json.NewEncoder(os.Stdout).Encode(response)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}
}
