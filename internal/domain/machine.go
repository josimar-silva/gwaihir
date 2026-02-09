// Package domain contains core business entities and interfaces.
package domain

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
)

// Machine represents a network machine that can be woken via WoL.
type Machine struct {
	ID        string `yaml:"id" json:"id"`
	Name      string `yaml:"name" json:"name"`
	MAC       string `yaml:"mac" json:"mac"`
	Broadcast string `yaml:"broadcast" json:"broadcast"`
}

// Validate checks if the machine has valid configuration.
func (m *Machine) Validate() error {
	if m.ID == "" {
		return errors.New("machine ID cannot be empty")
	}
	if m.Name == "" {
		return errors.New("machine name cannot be empty")
	}
	if err := ValidateMAC(m.MAC); err != nil {
		return fmt.Errorf("invalid MAC address: %w", err)
	}
	if err := ValidateBroadcast(m.Broadcast); err != nil {
		return fmt.Errorf("invalid broadcast address: %w", err)
	}
	return nil
}

// ValidateMAC validates a MAC address format.
func ValidateMAC(mac string) error {
	// Common MAC address formats: AA:BB:CC:DD:EE:FF or AA-BB-CC-DD-EE-FF
	macRegex := regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	if !macRegex.MatchString(mac) {
		return fmt.Errorf("MAC address must be in format XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX")
	}
	return nil
}

// ValidateBroadcast validates a broadcast IP address.
func ValidateBroadcast(broadcast string) error {
	ip := net.ParseIP(broadcast)
	if ip == nil {
		return fmt.Errorf("invalid IP address format")
	}
	// Check if it's an IPv4 address
	if ip.To4() == nil {
		return fmt.Errorf("broadcast address must be IPv4")
	}
	return nil
}

// NormalizeMAC normalizes a MAC address to use colons as separators.
func (m *Machine) NormalizeMAC() string {
	return strings.ReplaceAll(strings.ToUpper(m.MAC), "-", ":")
}
