package validator

import (
	"github.com/concourse/concourse-pipeline-resource/concourse"
)

func ValidateCheck(input concourse.CheckRequest) error  {
	return ValidateTeams(input.Source.Teams)
}
