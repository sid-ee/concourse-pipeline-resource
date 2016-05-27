package validator

import (
	"fmt"

	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
)

func ValidateCheck(input concourse.CheckRequest) error {
	if input.Source.Username == "" {
		return fmt.Errorf("%s must be provided", "username")
	}

	if input.Source.Password == "" {
		return fmt.Errorf("%s must be provided", "password")
	}

	return nil
}
