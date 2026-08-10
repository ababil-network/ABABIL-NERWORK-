package app

import (
	"errors"
	"time"
)

func parseBlockTimestamp(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("invalid block timestamp")
	}

	t, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, errors.New("invalid block timestamp")
	}

	return t.UTC(), nil
}
