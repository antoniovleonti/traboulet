package kuba

import (
	"errors"
	"time"
)

type Config struct {
	TimeControl time.Duration `json:"timeControlNs"`
	// More config can go here in the future.
}

func (c *Config) Validate() error {
	if c.TimeControl <= 0 || c.TimeControl > time.Hour {
		return errors.New("time control should be > 0s and <= 1hr")
	}
	return nil
}
