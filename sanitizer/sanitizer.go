package sanitizer

import (
	"io"
	"strings"
)

type Sanitizer struct {
	sanitized map[string]string
	sink      io.Writer
}

func NewSanitizer(sanitized map[string]string, sink io.Writer) *Sanitizer {
	return &Sanitizer{
		sanitized: sanitized,
		sink:      sink,
	}
}

func (s Sanitizer) Write(p []byte) (n int, err error) {
	input := string(p)

	for k, v := range s.sanitized {
		input = strings.Replace(input, k, v, -1)
	}

	scrubbed := []byte(input)

	return s.sink.Write(scrubbed)
}
