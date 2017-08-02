package validator

import (
	"github.com/concourse/concourse-pipeline-resource/concourse"
)

func ValidateIn(input concourse.InRequest) error {
	return ValidateTeams(input.Source.Teams)
}
