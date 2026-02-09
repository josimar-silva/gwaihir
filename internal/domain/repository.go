package domain

// MachineRepository defines the interface for machine data access.
type MachineRepository interface {
	// GetByID retrieves a machine by its ID.
	GetByID(id string) (*Machine, error)

	// GetAll retrieves all registered machines.
	GetAll() ([]*Machine, error)

	// Exists checks if a machine with the given ID exists.
	Exists(id string) bool
}

// WoLPacketSender defines the interface for sending Wake-on-LAN magic packets.
type WoLPacketSender interface {
	// SendMagicPacket sends a WoL magic packet to the specified MAC on the broadcast address.
	SendMagicPacket(mac, broadcast string) error
}
