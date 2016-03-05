package concourse

func SanitizedSource(source Source) map[string]string {
	s := make(map[string]string)

	if source.Password != "" {
		s[source.Password] = "***REDACTED-PASSWORD***"
	}

	return s
}
