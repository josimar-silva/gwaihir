package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/josimar-silva/gwaihir/internal/config"
)

// Test 3.1.1: NewInMemoryMachineRepository accepts config struct
func TestNewInMemoryMachineRepository_WithConfig(t *testing.T) {
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
	repo, err := NewInMemoryMachineRepository(cfg)

	// Assert
	if err != nil {
		t.Fatalf("NewInMemoryMachineRepository() error = %v", err)
	}

	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}

	if len(repo.machines) != 2 {
		t.Errorf("Expected 2 machines, got %d", len(repo.machines))
	}
}

// Test 3.1.2: Machines loaded from config.Machines array
func TestNewInMemoryMachineRepository_LoadsFromConfig(t *testing.T) {
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
	repo, err := NewInMemoryMachineRepository(cfg)

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
func TestNewInMemoryMachineRepository_ValidatesMachines(t *testing.T) {
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
	_, err := NewInMemoryMachineRepository(cfg)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid machine")
	}
}

// Test 3.1.4: Error handling for duplicate IDs
func TestNewInMemoryMachineRepository_DuplicateID(t *testing.T) {
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
	_, err := NewInMemoryMachineRepository(cfg)

	// Assert
	if err == nil {
		t.Error("Expected error for duplicate machine ID")
	}
}

// Test 3.1.5: GetByID works with config machines
func TestInMemoryMachineRepository_GetByID(t *testing.T) {
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

	repo, err := NewInMemoryMachineRepository(cfg)
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
func TestInMemoryMachineRepository_GetAll(t *testing.T) {
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

	repo, err := NewInMemoryMachineRepository(cfg)
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
func TestInMemoryMachineRepository_Exists(t *testing.T) {
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

	repo, err := NewInMemoryMachineRepository(cfg)
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

// ============================================================================
// Integration Tests: Repository with Example Config
// ============================================================================

// Test 3.2.1: Load example config file and create repository
func TestInMemoryMachineRepository_IntegrationWithExampleConfig(t *testing.T) {
	// Arrange
	projectRoot := findProjectRoot(t)
	exampleConfigPath := filepath.Join(projectRoot, "configs", "gwaihir.example.yaml")

	// Act
	cfg, err := config.LoadConfig(exampleConfigPath)
	if err != nil {
		t.Fatalf("Failed to load example config: %v", err)
	}

	repo, err := NewInMemoryMachineRepository(cfg)

	// Assert
	if err != nil {
		t.Fatalf("Failed to create repository from config: %v", err)
	}

	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}

	// Verify machines loaded from config
	machines, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(machines) != 3 {
		t.Errorf("Expected 3 machines from example config, got %d", len(machines))
	}
}

// Test 3.2.2: Verify all machines from example config are valid
func TestInMemoryMachineRepository_IntegrationExampleMachinesValid(t *testing.T) {
	// Arrange
	projectRoot := findProjectRoot(t)
	exampleConfigPath := filepath.Join(projectRoot, "configs", "gwaihir.example.yaml")

	cfg, err := config.LoadConfig(exampleConfigPath)
	if err != nil {
		t.Fatalf("Failed to load example config: %v", err)
	}

	// Act
	repo, err := NewInMemoryMachineRepository(cfg)

	// Assert
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Each machine should pass domain validation
	expectedMachines := []string{"saruman", "gandalf", "radagast"}
	for _, machineID := range expectedMachines {
		machine, err := repo.GetByID(machineID)
		if err != nil {
			t.Errorf("GetByID(%s) error = %v", machineID, err)
			continue
		}

		if machine == nil {
			t.Errorf("Expected machine %s to be found", machineID)
			continue
		}

		if machine.ID != machineID {
			t.Errorf("Expected machine ID %s, got %s", machineID, machine.ID)
		}

		if machine.Name == "" {
			t.Errorf("Machine %s: name should not be empty", machineID)
		}

		if machine.MAC == "" {
			t.Errorf("Machine %s: MAC should not be empty", machineID)
		}

		if machine.Broadcast == "" {
			t.Errorf("Machine %s: broadcast should not be empty", machineID)
		}
	}
}

// Test 3.2.3: GetByID with example config machines
func TestInMemoryMachineRepository_IntegrationGetByIDWithExampleConfig(t *testing.T) {
	// Arrange
	projectRoot := findProjectRoot(t)
	exampleConfigPath := filepath.Join(projectRoot, "configs", "gwaihir.example.yaml")

	cfg, err := config.LoadConfig(exampleConfigPath)
	if err != nil {
		t.Fatalf("Failed to load example config: %v", err)
	}

	repo, err := NewInMemoryMachineRepository(cfg)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	tests := []struct {
		name         string
		machineID    string
		shouldFind   bool
		expectedName string
	}{
		{
			name:         "existing machine: saruman",
			machineID:    "saruman",
			shouldFind:   true,
			expectedName: "Development Server",
		},
		{
			name:         "existing machine: gandalf",
			machineID:    "gandalf",
			shouldFind:   true,
			expectedName: "Production Server",
		},
		{
			name:         "existing machine: radagast",
			machineID:    "radagast",
			shouldFind:   true,
			expectedName: "Backup Server",
		},
		{
			name:       "non-existing machine",
			machineID:  "nonexistent",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			machine, err := repo.GetByID(tt.machineID)

			if tt.shouldFind {
				if err != nil {
					t.Errorf("GetByID() error = %v", err)
				}
				if machine == nil {
					t.Error("Expected non-nil machine")
					return
				}
				if machine.Name != tt.expectedName {
					t.Errorf("Expected name %q, got %q", tt.expectedName, machine.Name)
				}
			} else if err == nil {
				t.Error("Expected error for non-existing machine")
			}
		})
	}
}

// Test 3.2.4: GetAll with example config
func TestInMemoryMachineRepository_IntegrationGetAllWithExampleConfig(t *testing.T) {
	// Arrange
	projectRoot := findProjectRoot(t)
	exampleConfigPath := filepath.Join(projectRoot, "configs", "gwaihir.example.yaml")

	cfg, err := config.LoadConfig(exampleConfigPath)
	if err != nil {
		t.Fatalf("Failed to load example config: %v", err)
	}

	repo, err := NewInMemoryMachineRepository(cfg)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Act
	machines, err := repo.GetAll()

	// Assert
	if err != nil {
		t.Errorf("GetAll() error = %v", err)
	}

	if len(machines) != 3 {
		t.Errorf("Expected 3 machines, got %d", len(machines))
	}

	// Verify all expected machines are present
	machineMap := make(map[string]bool)
	for _, m := range machines {
		machineMap[m.ID] = true
	}

	expectedIDs := []string{"saruman", "gandalf", "radagast"}
	for _, expectedID := range expectedIDs {
		if !machineMap[expectedID] {
			t.Errorf("Expected machine %s not found in GetAll()", expectedID)
		}
	}
}

// Helper function: findProjectRoot walks up the directory tree to find go.mod
func findProjectRoot(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			t.Fatalf("Could not find project root (go.mod). Current directory: %s", cwd)
		}
		cwd = parent
	}
}
