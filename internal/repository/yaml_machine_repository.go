// Package repository provides data access implementations.
package repository

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/josimar-silva/gwaihir/internal/domain"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
)

// Config represents the YAML configuration file structure.
type Config struct {
	Machines []domain.Machine `yaml:"machines"`
}

// YAMLMachineRepository implements MachineRepository using a YAML file.
type YAMLMachineRepository struct {
	machines map[string]*domain.Machine
	mu       sync.RWMutex
}

// NewYAMLMachineRepository creates a new YAML-based machine repository.
func NewYAMLMachineRepository(configPath string) (*YAMLMachineRepository, error) {
	// #nosec G304 - configPath is provided by the application configuration, not user input
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and index machines
	machines := make(map[string]*domain.Machine)
	for i := range config.Machines {
		machine := &config.Machines[i]
		if err := machine.Validate(); err != nil {
			return nil, fmt.Errorf("invalid machine %s: %w", machine.ID, err)
		}
		if _, exists := machines[machine.ID]; exists {
			return nil, fmt.Errorf("duplicate machine ID: %s", machine.ID)
		}
		machines[machine.ID] = machine
	}

	return &YAMLMachineRepository{
		machines: machines,
	}, nil
}

// GetByID retrieves a machine by its ID.
func (r *YAMLMachineRepository) GetByID(id string) (*domain.Machine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	machine, exists := r.machines[id]
	if !exists {
		return nil, domain.ErrMachineNotFound
	}

	return machine, nil
}

// GetAll retrieves all registered machines.
func (r *YAMLMachineRepository) GetAll() ([]*domain.Machine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	machines := make([]*domain.Machine, 0, len(r.machines))
	for _, machine := range r.machines {
		machines = append(machines, machine)
	}

	return machines, nil
}

// Exists checks if a machine with the given ID exists.
func (r *YAMLMachineRepository) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.machines[id]
	return exists
}

// NewWoLPacketSender creates a new WoL packet sender instance.
func NewWoLPacketSender() domain.WoLPacketSender {
	return infrastructure.NewWoLPacketSender()
}
