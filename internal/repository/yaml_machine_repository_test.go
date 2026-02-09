package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewYAMLMachineRepository(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "machines.yaml")

	validConfig := `machines:
  - id: server1
    name: Test Server 1
    mac: "AA:BB:CC:DD:EE:FF"
    broadcast: "192.168.1.255"
  - id: server2
    name: Test Server 2
    mac: "11:22:33:44:55:66"
    broadcast: "192.168.1.255"
`

	if err := os.WriteFile(configPath, []byte(validConfig), 0o600); err != nil {
		t.Fatal(err)
	}

	repo, err := NewYAMLMachineRepository(configPath)
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

func TestNewYAMLMachineRepository_InvalidFile(t *testing.T) {
	_, err := NewYAMLMachineRepository("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestNewYAMLMachineRepository_DuplicateID(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "machines.yaml")

	duplicateConfig := `machines:
  - id: server1
    name: Test Server 1
    mac: "AA:BB:CC:DD:EE:FF"
    broadcast: "192.168.1.255"
  - id: server1
    name: Test Server 2
    mac: "11:22:33:44:55:66"
    broadcast: "192.168.1.255"
`

	if err := os.WriteFile(configPath, []byte(duplicateConfig), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := NewYAMLMachineRepository(configPath)
	if err == nil {
		t.Error("Expected error for duplicate machine ID")
	}
}

func TestYAMLMachineRepository_GetByID(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "machines.yaml")

	config := `machines:
  - id: server1
    name: Test Server
    mac: "AA:BB:CC:DD:EE:FF"
    broadcast: "192.168.1.255"
`

	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	repo, err := NewYAMLMachineRepository(configPath)
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

func TestYAMLMachineRepository_GetAll(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "machines.yaml")

	config := `machines:
  - id: server1
    name: Test Server 1
    mac: "AA:BB:CC:DD:EE:FF"
    broadcast: "192.168.1.255"
  - id: server2
    name: Test Server 2
    mac: "11:22:33:44:55:66"
    broadcast: "192.168.1.255"
`

	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	repo, err := NewYAMLMachineRepository(configPath)
	if err != nil {
		t.Fatal(err)
	}

	machines, err := repo.GetAll()
	if err != nil {
		t.Errorf("GetAll() error = %v", err)
	}

	if len(machines) != 2 {
		t.Errorf("Expected 2 machines, got %d", len(machines))
	}
}

func TestYAMLMachineRepository_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "machines.yaml")

	config := `machines:
  - id: server1
    name: Test Server
    mac: "AA:BB:CC:DD:EE:FF"
    broadcast: "192.168.1.255"
`

	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}

	repo, err := NewYAMLMachineRepository(configPath)
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
