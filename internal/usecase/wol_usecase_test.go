package usecase

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/josimar-silva/gwaihir/internal/domain"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
)

// Mock implementations for testing

type mockMachineRepository struct {
	machines     map[string]*domain.Machine
	getByIDError error
	getAllError  error
}

func newMockMachineRepository(machines map[string]*domain.Machine) *mockMachineRepository {
	return &mockMachineRepository{
		machines: machines,
	}
}

func (m *mockMachineRepository) GetByID(id string) (*domain.Machine, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	machine, ok := m.machines[id]
	if !ok {
		return nil, domain.ErrMachineNotFound
	}
	return machine, nil
}

func (m *mockMachineRepository) GetAll() ([]*domain.Machine, error) {
	if m.getAllError != nil {
		return nil, m.getAllError
	}
	machines := make([]*domain.Machine, 0, len(m.machines))
	for _, machine := range m.machines {
		machines = append(machines, machine)
	}
	return machines, nil
}

func (m *mockMachineRepository) Exists(id string) bool {
	_, ok := m.machines[id]
	return ok
}

type mockWoLPacketSender struct {
	sendPackets    []sentPacket
	sendError      error
	sendErrorCount int
	callCount      int
}

type sentPacket struct {
	mac       string
	broadcast string
}

func newMockWoLPacketSender() *mockWoLPacketSender {
	return &mockWoLPacketSender{
		sendPackets: make([]sentPacket, 0),
	}
}

func (m *mockWoLPacketSender) SendMagicPacket(mac, broadcast string) error {
	m.callCount++
	m.sendPackets = append(m.sendPackets, sentPacket{mac: mac, broadcast: broadcast})

	if m.sendErrorCount > 0 {
		m.sendErrorCount--
		return m.sendError
	}
	return nil
}

// Test cases

func TestSendWakePacket_Success(t *testing.T) {
	// Arrange
	machines := map[string]*domain.Machine{
		"saruman": {
			ID:        "saruman",
			Name:      "Saruman Server",
			MAC:       "AA:BB:CC:DD:EE:FF",
			Broadcast: "192.168.1.255",
		},
	}
	repo := newMockMachineRepository(machines)
	sender := newMockWoLPacketSender()
	logger := newTestLogger()
	metrics := newTestMetrics()
	useCase := NewWoLUseCase(repo, sender, logger, metrics)

	// Act
	err := useCase.SendWakePacket("saruman")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if sender.callCount != 1 {
		t.Errorf("Expected SendMagicPacket to be called once, got %d", sender.callCount)
	}
	if len(sender.sendPackets) != 1 {
		t.Errorf("Expected 1 packet sent, got %d", len(sender.sendPackets))
	}
	if sender.sendPackets[0].mac != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("Expected MAC AA:BB:CC:DD:EE:FF, got %s", sender.sendPackets[0].mac)
	}
	if sender.sendPackets[0].broadcast != "192.168.1.255" {
		t.Errorf("Expected broadcast 192.168.1.255, got %s", sender.sendPackets[0].broadcast)
	}
}

func TestSendWakePacket_MachineNotFound(t *testing.T) {
	// Arrange
	repo := newMockMachineRepository(map[string]*domain.Machine{})
	sender := newMockWoLPacketSender()
	logger := newTestLogger()
	metrics := newTestMetrics()
	useCase := NewWoLUseCase(repo, sender, logger, metrics)

	// Act
	err := useCase.SendWakePacket("nonexistent")

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !errors.Is(err, domain.ErrMachineNotFound) {
		t.Errorf("Expected ErrMachineNotFound, got %v", err)
	}
	if sender.callCount != 0 {
		t.Errorf("Expected SendMagicPacket not to be called, got %d calls", sender.callCount)
	}
}

func TestSendWakePacket_SendError(t *testing.T) {
	// Arrange
	machines := map[string]*domain.Machine{
		"saruman": {
			ID:        "saruman",
			Name:      "Saruman Server",
			MAC:       "AA:BB:CC:DD:EE:FF",
			Broadcast: "192.168.1.255",
		},
	}
	repo := newMockMachineRepository(machines)
	sender := newMockWoLPacketSender()
	sender.sendError = errors.New("network error")
	sender.sendErrorCount = 1
	logger := newTestLogger()
	metrics := newTestMetrics()

	useCase := NewWoLUseCase(repo, sender, logger, metrics)

	// Act
	err := useCase.SendWakePacket("saruman")

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !contains(err.Error(), "failed to send WoL packet") {
		t.Errorf("Expected error to contain 'failed to send WoL packet', got %s", err.Error())
	}
}

func TestSendWakePacket_MultipleMachines(t *testing.T) {
	// Arrange
	machines := map[string]*domain.Machine{
		"saruman": {
			ID:        "saruman",
			Name:      "Saruman Server",
			MAC:       "AA:BB:CC:DD:EE:FF",
			Broadcast: "192.168.1.255",
		},
		"morgoth": {
			ID:        "morgoth",
			Name:      "Morgoth Server",
			MAC:       "11:22:33:44:55:66",
			Broadcast: "192.168.1.255",
		},
	}
	repo := newMockMachineRepository(machines)
	sender := newMockWoLPacketSender()
	logger := newTestLogger()
	metrics := newTestMetrics()
	useCase := NewWoLUseCase(repo, sender, logger, metrics)

	// Act
	err1 := useCase.SendWakePacket("saruman")
	err2 := useCase.SendWakePacket("morgoth")

	// Assert
	if err1 != nil || err2 != nil {
		t.Errorf("Expected no errors, got %v and %v", err1, err2)
	}
	if sender.callCount != 2 {
		t.Errorf("Expected 2 packets sent, got %d", sender.callCount)
	}
	if sender.sendPackets[0].mac != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("First packet: expected MAC AA:BB:CC:DD:EE:FF, got %s", sender.sendPackets[0].mac)
	}
	if sender.sendPackets[1].mac != "11:22:33:44:55:66" {
		t.Errorf("Second packet: expected MAC 11:22:33:44:55:66, got %s", sender.sendPackets[1].mac)
	}
}

func TestListMachines_Success(t *testing.T) {
	// Arrange
	machines := map[string]*domain.Machine{
		"saruman": {
			ID:   "saruman",
			Name: "Saruman Server",
			MAC:  "AA:BB:CC:DD:EE:FF",
		},
		"morgoth": {
			ID:   "morgoth",
			Name: "Morgoth Server",
			MAC:  "11:22:33:44:55:66",
		},
	}
	repo := newMockMachineRepository(machines)
	sender := newMockWoLPacketSender()
	logger := newTestLogger()
	metrics := newTestMetrics()
	useCase := NewWoLUseCase(repo, sender, logger, metrics)

	// Act
	result, err := useCase.ListMachines()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 machines, got %d", len(result))
	}
}

func TestListMachines_Error(t *testing.T) {
	// Arrange
	repo := newMockMachineRepository(nil)
	repo.getAllError = errors.New("database error")
	sender := newMockWoLPacketSender()
	logger := newTestLogger()
	metrics := newTestMetrics()
	useCase := NewWoLUseCase(repo, sender, logger, metrics)

	// Act
	result, err := useCase.ListMachines()

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

func TestListMachines_Empty(t *testing.T) {
	// Arrange
	repo := newMockMachineRepository(map[string]*domain.Machine{})
	sender := newMockWoLPacketSender()
	logger := newTestLogger()
	metrics := newTestMetrics()
	useCase := NewWoLUseCase(repo, sender, logger, metrics)

	// Act
	result, err := useCase.ListMachines()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 machines, got %d", len(result))
	}
}

func TestGetMachine_Success(t *testing.T) {
	// Arrange
	machines := map[string]*domain.Machine{
		"saruman": {
			ID:        "saruman",
			Name:      "Saruman Server",
			MAC:       "AA:BB:CC:DD:EE:FF",
			Broadcast: "192.168.1.255",
		},
	}
	repo := newMockMachineRepository(machines)
	sender := newMockWoLPacketSender()
	logger := newTestLogger()
	metrics := newTestMetrics()
	useCase := NewWoLUseCase(repo, sender, logger, metrics)

	// Act
	machine, err := useCase.GetMachine("saruman")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if machine == nil {
		t.Fatal("Expected machine, got nil")
	}
	if machine.ID != "saruman" {
		t.Errorf("Expected ID saruman, got %s", machine.ID)
	}
	if machine.Name != "Saruman Server" {
		t.Errorf("Expected name 'Saruman Server', got %s", machine.Name)
	}
}

func TestGetMachine_NotFound(t *testing.T) {
	// Arrange
	repo := newMockMachineRepository(map[string]*domain.Machine{})
	sender := newMockWoLPacketSender()
	logger := newTestLogger()
	metrics := newTestMetrics()
	useCase := NewWoLUseCase(repo, sender, logger, metrics)

	// Act
	machine, err := useCase.GetMachine("nonexistent")

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !errors.Is(err, domain.ErrMachineNotFound) {
		t.Errorf("Expected ErrMachineNotFound, got %v", err)
	}
	if machine != nil {
		t.Errorf("Expected nil machine, got %v", machine)
	}
}

func TestNewWoLUseCase(t *testing.T) {
	repo := newMockMachineRepository(nil)
	sender := newMockWoLPacketSender()
	logger := newTestLogger()
	metrics := newTestMetrics()

	useCase := NewWoLUseCase(repo, sender, logger, metrics)

	if useCase == nil {
		t.Fatal("Expected non-nil usecase")
	}
	if useCase.machineRepo != repo {
		t.Error("Expected machineRepo to be set")
	}
	if useCase.packetSender != sender {
		t.Error("Expected packetSender to be set")
	}
	if useCase.logger != logger {
		t.Error("Expected logger to be set")
	}
	if useCase.metrics != metrics {
		t.Error("Expected metrics to be set")
	}
}

// Helper functions
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func newTestLogger() *infrastructure.Logger {
	return infrastructure.NewLogger(false)
}

func newTestMetrics() *infrastructure.Metrics {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	metrics, _ := infrastructure.NewMetrics()
	return metrics
}
