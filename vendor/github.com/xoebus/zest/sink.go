package zest

import (
	"encoding/json"
	"errors"

	"github.com/contraband/holler"
	"github.com/pivotal-golang/lager"
)

// Sink is the type that represents the sink that will emit errors to Yeller.
type Sink struct {
	yeller *holler.Yeller
}

// NewYellerSink creates a new Sink for use with Lager.
func NewYellerSink(token, env string) *Sink {
	return &Sink{
		yeller: holler.NewYeller(token, env),
	}
}

// Log will send any error log lines up to Yeller.
func (s *Sink) Log(level lager.LogLevel, message []byte) {
	if level < lager.ERROR {
		return
	}

	var line lager.LogFormat
	if err := json.Unmarshal(message, &line); err != nil {
		return
	}

	if errStr, ok := line.Data["error"].(string); ok {
		delete(line.Data, "message")
		delete(line.Data, "error")

		s.yeller.Notify(
			line.Message,
			errors.New(errStr),
			line.Data,
		)
	}
}
