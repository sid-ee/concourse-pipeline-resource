package concourse

import "fmt"

func SanitizedSource(source Source) map[string]string {
	s := make(map[string]string)

	for i, t := range source.Teams {
		if t.Password != "" {
			s[t.Password] = fmt.Sprintf("***REDACTED-PASSWORD-TEAM-%d***", i)
		}
	}

	return s
}
