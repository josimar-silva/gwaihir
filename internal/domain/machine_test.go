package domain

import (
	"testing"
)

func TestMachine_Validate(t *testing.T) {
	tests := []struct {
		name    string
		machine Machine
		wantErr bool
	}{
		{
			name: "valid machine",
			machine: Machine{
				ID:        "server1",
				Name:      "Test Server",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
			wantErr: false,
		},
		{
			name: "valid machine with dash separator",
			machine: Machine{
				ID:        "server2",
				Name:      "Test Server 2",
				MAC:       "AA-BB-CC-DD-EE-FF",
				Broadcast: "192.168.1.255",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			machine: Machine{
				Name:      "Test Server",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			machine: Machine{
				ID:        "server1",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
			wantErr: true,
		},
		{
			name: "invalid MAC",
			machine: Machine{
				ID:        "server1",
				Name:      "Test Server",
				MAC:       "invalid",
				Broadcast: "192.168.1.255",
			},
			wantErr: true,
		},
		{
			name: "invalid broadcast",
			machine: Machine{
				ID:        "server1",
				Name:      "Test Server",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.machine.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Machine.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMachine_NormalizeMAC(t *testing.T) {
	tests := []struct {
		name string
		mac  string
		want string
	}{
		{
			name: "colon separator",
			mac:  "aa:bb:cc:dd:ee:ff",
			want: "AA:BB:CC:DD:EE:FF",
		},
		{
			name: "dash separator",
			mac:  "aa-bb-cc-dd-ee-ff",
			want: "AA:BB:CC:DD:EE:FF",
		},
		{
			name: "mixed case",
			mac:  "Aa:Bb:Cc:Dd:Ee:Ff",
			want: "AA:BB:CC:DD:EE:FF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Machine{MAC: tt.mac}
			if got := m.NormalizeMAC(); got != tt.want {
				t.Errorf("Machine.NormalizeMAC() = %v, want %v", got, tt.want)
			}
		})
	}
}
