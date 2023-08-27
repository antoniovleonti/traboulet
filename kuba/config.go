package kuba

import "time"

type Config struct {
	TimeControl time.Duration `json:"TimeControlNs"`
	// More config can go here in the future.
}
