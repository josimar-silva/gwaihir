// Package repository provides data access implementations.
package repository

import (
	"fmt"
	"sync"

	"github.com/josimar-silva/gwaihir/internal/config"
	"github.com/josimar-silva/gwaihir/internal/domain"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
)

// YAMLMachineRepository implements MachineRepository using configuration.
type YAMLMachineRepository struct {
	machines map[string]*domain.Machine
	mu       sync.RWMutex
}

// NewYAMLMachineRepository creates a new machine repository from config.
func NewYAMLMachineRepository(cfg *config.Config) (*YAMLMachineRepository, error) {
	// Build and validate machines from config
	machines := make(map[string]*domain.Machine)
	for i := range cfg.Machines {
		machineConfig := &cfg.Machines[i]

		// Convert config.MachineConfig to domain.Machine
		machine := &domain.Machine{
			ID:        machineConfig.ID,
			Name:      machineConfig.Name,
			MAC:       machineConfig.MAC,
			Broadcast: machineConfig.Broadcast,
		}

		// Validate using domain validation
		if err := machine.Validate(); err != nil {
			return nil, fmt.Errorf("invalid machine %s: %w", machine.ID, err)
		}

		// Check for duplicates
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
