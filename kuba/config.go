package kuba

import "time"

type Config struct {
	InitialTime time.Duration `json:"initialTimeNs"`
	// More config can go here in the future.
}
