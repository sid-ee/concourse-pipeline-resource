package validator

import (
	"fmt"

	"github.com/concourse/concourse-pipeline-resource/concourse"
)

func ValidateCheck(input concourse.CheckRequest) error {
	if input.Source.Teams == nil {
		return fmt.Errorf("%s must be provided in source", "teams")
	}

	for i, team := range input.Source.Teams {
		if team.Name == "" {
			return fmt.Errorf("%s must be provided for team: %d", "name", i)
		}

		if team.Username == "" {
			return fmt.Errorf("%s must be provided for team: %s", "username", team.Name)
		}

		if team.Password == "" {
			return fmt.Errorf("%s must be provided for team: %s", "password", team.Name)
		}
	}

	return nil
}
