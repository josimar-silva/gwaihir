package domain

import "errors"

// Domain errors.
var (
	// ErrMachineNotFound is returned when a requested machine is not found.
	ErrMachineNotFound = errors.New("machine not found")

	// ErrMachineNotAllowed is returned when a machine is not in the allowlist.
	ErrMachineNotAllowed = errors.New("machine not allowed")

	// ErrInvalidConfiguration is returned when configuration is invalid.
	ErrInvalidConfiguration = errors.New("invalid configuration")
)
