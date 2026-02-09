// Package usecase contains business logic use cases.
package usecase

import (
	"fmt"

	"github.com/josimar-silva/gwaihir/internal/domain"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
)

// WoLUseCase handles the business logic for sending WoL packets.
type WoLUseCase struct {
	machineRepo  domain.MachineRepository
	packetSender domain.WoLPacketSender
	logger       *infrastructure.Logger
	metrics      *infrastructure.Metrics
}

// NewWoLUseCase creates a new WoL use case.
func NewWoLUseCase(machineRepo domain.MachineRepository, packetSender domain.WoLPacketSender, logger *infrastructure.Logger, metrics *infrastructure.Metrics) *WoLUseCase {
	return &WoLUseCase{
		machineRepo:  machineRepo,
		packetSender: packetSender,
		logger:       logger,
		metrics:      metrics,
	}
}

// SendWakePacket sends a WoL packet to the specified machine.
// It validates that the machine is in the allowlist before sending.
func (uc *WoLUseCase) SendWakePacket(machineID string) error {
	machine, err := uc.machineRepo.GetByID(machineID)
	if err != nil {
		uc.metrics.MachineNotFound.Inc()
		return fmt.Errorf("failed to get machine: %w", err)
	}

	uc.logger.Info("Sending WoL packet",
		infrastructure.String("machine_id", machine.ID),
		infrastructure.String("machine_name", machine.Name),
		infrastructure.String("mac", machine.NormalizeMAC()),
		infrastructure.String("broadcast", machine.Broadcast),
	)

	if err := uc.packetSender.SendMagicPacket(machine.MAC, machine.Broadcast); err != nil {
		uc.metrics.WoLPacketsFailed.Inc()
		uc.logger.Error("Failed to send WoL packet",
			infrastructure.String("machine_id", machine.ID),
			infrastructure.Any("error", err),
		)
		return fmt.Errorf("failed to send WoL packet: %w", err)
	}

	uc.metrics.WoLPacketsSent.Inc()
	uc.logger.Info("WoL packet sent successfully",
		infrastructure.String("machine_id", machine.ID),
	)
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
