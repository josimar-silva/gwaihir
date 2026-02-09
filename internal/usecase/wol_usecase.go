// Package usecase contains business logic use cases.
package usecase

import (
	"fmt"
	"log"

	"github.com/josimar-silva/gwaihir/internal/domain"
)

// WoLUseCase handles the business logic for sending WoL packets.
type WoLUseCase struct {
	machineRepo  domain.MachineRepository
	packetSender domain.WoLPacketSender
}

// NewWoLUseCase creates a new WoL use case.
func NewWoLUseCase(machineRepo domain.MachineRepository, packetSender domain.WoLPacketSender) *WoLUseCase {
	return &WoLUseCase{
		machineRepo:  machineRepo,
		packetSender: packetSender,
	}
}

// SendWakePacket sends a WoL packet to the specified machine.
// It validates that the machine is in the allowlist before sending.
func (uc *WoLUseCase) SendWakePacket(machineID string) error {
	machine, err := uc.machineRepo.GetByID(machineID)
	if err != nil {
		return fmt.Errorf("failed to get machine: %w", err)
	}

	log.Printf("Sending WoL packet to machine '%s' (%s) at MAC %s on broadcast %s",
		machine.Name, machine.ID, machine.NormalizeMAC(), machine.Broadcast)

	if err := uc.packetSender.SendMagicPacket(machine.MAC, machine.Broadcast); err != nil {
		return fmt.Errorf("failed to send WoL packet: %w", err)
	}

	log.Printf("WoL packet successfully sent to machine '%s'", machine.ID)
	return nil
}

// ListMachines returns all registered machines.
func (uc *WoLUseCase) ListMachines() ([]*domain.Machine, error) {
	return uc.machineRepo.GetAll()
}

// GetMachine retrieves a specific machine by ID.
func (uc *WoLUseCase) GetMachine(machineID string) (*domain.Machine, error) {
	return uc.machineRepo.GetByID(machineID)
}
