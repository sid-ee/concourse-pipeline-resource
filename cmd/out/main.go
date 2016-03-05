package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
	"github.com/robdimsdale/concourse-pipeline-resource/logger"
	"github.com/robdimsdale/concourse-pipeline-resource/out"
	"github.com/robdimsdale/concourse-pipeline-resource/sanitizer"
)

const (
	flyBinaryName = "fly"
)

var (
	// version is deliberately left uninitialized so it can be set at compile-time
	version string

	l logger.Logger
)

func main() {
	if version == "" {
		version = "dev"
	}

	outDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}

	flyBinaryPath := filepath.Join(outDir, flyBinaryName)

	var input concourse.OutRequest

	logFile, err := ioutil.TempFile("", "concourse-resource-out.log")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintf(logFile, "Concourse Pipeline Resource version: %s\n", version)

	fmt.Fprintf(os.Stderr, "Logging to %s\n", logFile.Name())

	err = json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		fmt.Fprintf(logFile, "Exiting with error: %v\n", err)
		log.Fatalln(err)
	}

	sanitized := concourse.SanitizedSource(input.Source)
	sanitizer := sanitizer.NewSanitizer(sanitized, logFile)

	l = logger.NewLogger(sanitizer)

	_, err = out.NewOutCommand(version, l, flyBinaryPath).Run(input)
	if err != nil {
		l.Debugf("Exiting with error: %v\n", err)
		log.Fatalln(err)
	}
}
