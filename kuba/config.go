package kuba

import "time"

type Config struct {
	InitialTime time.Duration `json:"initial_time"`
	// More config can go here in the future.
}
