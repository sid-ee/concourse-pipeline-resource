package validator

import (
	"fmt"

	"github.com/robdimsdale/concourse-pipeline-resource/concourse"
)

func ValidateOut(input concourse.OutRequest) error {
	if input.Source.Username == "" {
		return fmt.Errorf("%s must be provided", "username")
	}

	if input.Source.Password == "" {
		return fmt.Errorf("%s must be provided", "password")
	}

	var pipelinesFilePresent bool
	var pipelinesPresent bool

	if input.Params.PipelinesFile != "" {
		pipelinesFilePresent = true
	}

	if input.Params.Pipelines != nil && len(input.Params.Pipelines) > 0 {
		pipelinesPresent = true
	}

	if !(pipelinesPresent || pipelinesFilePresent) {
		return fmt.Errorf(
			"pipelines must be provided via either %s or %s",
			"pipelines",
			"pipelines_file",
		)
	}

	if pipelinesPresent && pipelinesFilePresent {
		return fmt.Errorf(
			"pipelines must be provided via one of either %s or %s",
			"pipelines",
			"pipelines_file",
		)
	}

	for i, p := range input.Params.Pipelines {
		if p.Name == "" {
			return fmt.Errorf("%s must be provided for pipeline[%d]", "name", i)
		}

		if p.ConfigFile == "" {
			return fmt.Errorf("%s must be provided for pipeline[%d]", "config_file", i)
		}

		// vars files can be nil as it is optional.
		if p.VarsFiles != nil {
			// However, if it is provided it must be non-empty
			if len(p.VarsFiles) == 0 {
				return fmt.Errorf("%s must be non-empty if provided for pipeline[%d]", "vars_files", i)
			}

			for j, v := range p.VarsFiles {
				if len(v) == 0 {
					return fmt.Errorf(
						"%s must be non-empty for pipeline[%d].vars_files[%d]",
						"vars file",
						i,
						j,
					)
				}
			}
		}
	}

	return nil
}
