package validator

import (
	"fmt"

	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
)

func ValidateIn(input concourse.InRequest) error {
	if input.Source.Username == "" {
		return fmt.Errorf("%s must be provided", "username")
	}

	if input.Source.Password == "" {
		return fmt.Errorf("%s must be provided", "password")
	}

	return nil
}
