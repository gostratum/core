package core

import "errors"

// Common errors for the core package.
var (
	// ErrConfigNotFound is returned when a configuration key is not found.
	ErrConfigNotFound = errors.New("config key not found")
	// ErrHealthCheckFailed is returned when a health check fails.
	ErrHealthCheckFailed = errors.New("health check failed")
)
