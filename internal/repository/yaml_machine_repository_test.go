package repository

import (
	"testing"

	"github.com/josimar-silva/gwaihir/internal/config"
)

// Test 3.1.1: NewYAMLMachineRepository accepts config struct
func TestNewYAMLMachineRepository_WithConfig(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "server1",
				Name:      "Test Server 1",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
			{
				ID:        "server2",
				Name:      "Test Server 2",
				MAC:       "11:22:33:44:55:66",
				Broadcast: "192.168.1.255",
			},
		},
	}

	// Act
	repo, err := NewYAMLMachineRepository(cfg)

	// Assert
	if err != nil {
		t.Fatalf("NewYAMLMachineRepository() error = %v", err)
	}

	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}

	if len(repo.machines) != 2 {
		t.Errorf("Expected 2 machines, got %d", len(repo.machines))
	}
}

// Test 3.1.2: Machines loaded from config.Machines array
func TestNewYAMLMachineRepository_LoadsFromConfig(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "machine1",
				Name:      "Machine One",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
		},
	}

	// Act
	repo, err := NewYAMLMachineRepository(cfg)

	// Assert
	if err != nil {
		t.Fatal(err)
	}

	machine, err := repo.GetByID("machine1")
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if machine.ID != "machine1" {
		t.Errorf("Expected machine ID 'machine1', got '%s'", machine.ID)
	}
	if machine.Name != "Machine One" {
		t.Errorf("Expected name 'Machine One', got '%s'", machine.Name)
	}
}

// Test 3.1.3: Machines validated using domain.Machine
func TestNewYAMLMachineRepository_ValidatesMachines(t *testing.T) {
	// Arrange: config with invalid machine (bad MAC)
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "server1",
				Name:      "Test Server",
				MAC:       "INVALID",
				Broadcast: "192.168.1.255",
			},
		},
	}

	// Act
	_, err := NewYAMLMachineRepository(cfg)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid machine")
	}
}

// Test 3.1.4: Error handling for duplicate IDs
func TestNewYAMLMachineRepository_DuplicateID(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "server1",
				Name:      "Server One",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
			{
				ID:        "server1",
				Name:      "Server Two",
				MAC:       "11:22:33:44:55:66",
				Broadcast: "192.168.1.255",
			},
		},
	}

	// Act
	_, err := NewYAMLMachineRepository(cfg)

	// Assert
	if err == nil {
		t.Error("Expected error for duplicate machine ID")
	}
}

// Test 3.1.5: GetByID works with config machines
func TestYAMLMachineRepository_GetByID(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "server1",
				Name:      "Test Server",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
		},
	}

	repo, err := NewYAMLMachineRepository(cfg)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "existing machine",
			id:      "server1",
			wantErr: false,
		},
		{
			name:    "non-existing machine",
			id:      "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine, err := repo.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && machine == nil {
				t.Error("Expected non-nil machine")
			}
			if !tt.wantErr && machine.ID != tt.id {
				t.Errorf("Expected machine ID %s, got %s", tt.id, machine.ID)
			}
		})
	}
}

// Test 3.1.6: GetAll works with config machines
func TestYAMLMachineRepository_GetAll(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "server1",
				Name:      "Test Server 1",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
			{
				ID:        "server2",
				Name:      "Test Server 2",
				MAC:       "11:22:33:44:55:66",
				Broadcast: "192.168.1.255",
			},
		},
	}

	repo, err := NewYAMLMachineRepository(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Act
	machines, err := repo.GetAll()

	// Assert
	if err != nil {
		t.Errorf("GetAll() error = %v", err)
	}

	if len(machines) != 2 {
		t.Errorf("Expected 2 machines, got %d", len(machines))
	}
}

// Test 3.1.7: Exists works with config machines
func TestYAMLMachineRepository_Exists(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "server1",
				Name:      "Test Server",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
		},
	}

	repo, err := NewYAMLMachineRepository(cfg)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "existing machine",
			id:   "server1",
			want: true,
		},
		{
			name: "non-existing machine",
			id:   "nonexistent",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repo.Exists(tt.id); got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}
