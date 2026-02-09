package infrastructure

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestBuildMagicPacket(t *testing.T) {
	tests := []struct {
		name     string
		mac      string
		expected int
	}{
		{
			name:     "valid MAC address",
			mac:      "AA:BB:CC:DD:EE:FF",
			expected: 102,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			macBytes, err := net.ParseMAC(tt.mac)
			if err != nil {
				t.Fatalf("Failed to parse MAC: %v", err)
			}

			packet := buildMagicPacket(macBytes)

			if len(packet) != tt.expected {
				t.Errorf("Expected packet size %d, got %d", tt.expected, len(packet))
			}

			// Verify first 6 bytes are 0xFF
			for i := 0; i < 6; i++ {
				if packet[i] != 0xFF {
					t.Errorf("Byte %d: expected 0xFF, got 0x%X", i, packet[i])
				}
			}

			// Verify MAC is repeated 16 times
			for i := 0; i < 16; i++ {
				offset := 6 + i*6
				for j := 0; j < 6; j++ {
					if packet[offset+j] != macBytes[j] {
						t.Errorf("Repetition %d byte %d: expected 0x%X, got 0x%X",
							i, j, macBytes[j], packet[offset+j])
					}
				}
			}
		})
	}
}

func TestNormalizeMACAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "colon separator lowercase",
			input:    "aa:bb:cc:dd:ee:ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "colon separator uppercase",
			input:    "AA:BB:CC:DD:EE:FF",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "colon separator mixed case",
			input:    "Aa:Bb:Cc:Dd:Ee:Ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "dash separator lowercase",
			input:    "aa-bb-cc-dd-ee-ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "dash separator uppercase",
			input:    "AA-BB-CC-DD-EE-FF",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "dash separator mixed case",
			input:    "Aa-Bb-Cc-Dd-Ee-Ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeMACAddress(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSendMagicPacketValidation(t *testing.T) {
	tests := []struct {
		name          string
		mac           string
		broadcast     string
		shouldFail    bool
		errorContains string
	}{
		{
			name:       "valid MAC and broadcast",
			mac:        "AA:BB:CC:DD:EE:FF",
			broadcast:  "192.168.1.255",
			shouldFail: false,
		},
		{
			name:          "invalid MAC format",
			mac:           "invalid-mac",
			broadcast:     "192.168.1.255",
			shouldFail:    true,
			errorContains: "invalid MAC address",
		},
		{
			name:          "invalid broadcast address",
			mac:           "AA:BB:CC:DD:EE:FF",
			broadcast:     "invalid-broadcast",
			shouldFail:    true,
			errorContains: "failed to resolve",
		},
		{
			name:       "dash separated MAC",
			mac:        "AA-BB-CC-DD-EE-FF",
			broadcast:  "192.168.1.255",
			shouldFail: false,
		},
		{
			name:       "lowercase MAC",
			mac:        "aa:bb:cc:dd:ee:ff",
			broadcast:  "192.168.1.255",
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDialer := func(_, _ string) (net.PacketConn, error) {
				return &mockPacketConn{
					writeToFunc: func(b []byte, _ net.Addr) (int, error) {
						return len(b), nil
					},
				}, nil
			}

			sender := NewWoLPacketSenderWithDialer(mockDialer)
			err := sender.SendMagicPacket(tt.mac, tt.broadcast)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestSendMagicPacketPacketContent(t *testing.T) {
	mac := "AA:BB:CC:DD:EE:FF"
	broadcast := "192.168.1.255"

	var capturedPacket []byte
	var capturedAddr net.Addr

	mockDialer := func(_, _ string) (net.PacketConn, error) {
		return &mockPacketConn{
			writeToFunc: func(b []byte, addr net.Addr) (int, error) {
				capturedPacket = make([]byte, len(b))
				copy(capturedPacket, b)
				capturedAddr = addr
				return len(b), nil
			},
		}, nil
	}

	sender := NewWoLPacketSenderWithDialer(mockDialer)
	err := sender.SendMagicPacket(mac, broadcast)

	if err != nil {
		t.Fatalf("Failed to send packet: %v", err)
	}

	if capturedPacket == nil {
		t.Fatal("No packet was captured")
	}

	if len(capturedPacket) != 102 {
		t.Errorf("Expected packet size 102, got %d", len(capturedPacket))
	}

	// Verify first 6 bytes are 0xFF
	for i := 0; i < 6; i++ {
		if capturedPacket[i] != 0xFF {
			t.Errorf("Byte %d: expected 0xFF, got 0x%X", i, capturedPacket[i])
		}
	}

	// Verify packet contains MAC 16 times
	macBytes, _ := net.ParseMAC(mac)
	for i := 0; i < 16; i++ {
		offset := 6 + i*6
		for j := 0; j < 6; j++ {
			if capturedPacket[offset+j] != macBytes[j] {
				t.Errorf("Repetition %d byte %d: expected 0x%X, got 0x%X",
					i, j, macBytes[j], capturedPacket[offset+j])
			}
		}
	}

	if capturedAddr.String() != "192.168.1.255:9" {
		t.Errorf("Expected address 192.168.1.255:9, got %s", capturedAddr.String())
	}
}

func TestSendMagicPacketNetworkError(t *testing.T) {
	tests := []struct {
		name          string
		dialError     error
		writeToError  error
		errorContains string
	}{
		{
			name:          "dial fails",
			dialError:     fmt.Errorf("network unavailable"),
			errorContains: "failed to create UDP connection",
		},
		{
			name:          "writeTo fails",
			writeToError:  fmt.Errorf("write failed"),
			errorContains: "failed to send magic packet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockDialer func(network, addr string) (net.PacketConn, error)

			if tt.dialError != nil {
				mockDialer = func(_, _ string) (net.PacketConn, error) {
					return nil, tt.dialError
				}
			} else {
				mockDialer = func(_, _ string) (net.PacketConn, error) {
					return &mockPacketConn{
						writeToFunc: func(_ []byte, _ net.Addr) (int, error) {
							return 0, tt.writeToError
						},
					}, nil
				}
			}

			sender := NewWoLPacketSenderWithDialer(mockDialer)
			err := sender.SendMagicPacket("AA:BB:CC:DD:EE:FF", "192.168.1.255")

			if err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !contains(err.Error(), tt.errorContains) {
				t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
			}
		})
	}
}

func TestNewWoLPacketSender(t *testing.T) {
	sender := NewWoLPacketSender()
	if sender == nil {
		t.Fatal("Expected non-nil sender")
	}
	if sender.dialFunc == nil {
		t.Fatal("Expected dialFunc to be set")
	}
}

func TestNewWoLPacketSenderWithDialer(t *testing.T) {
	customDialer := func(_, _ string) (net.PacketConn, error) {
		return nil, nil
	}

	sender := NewWoLPacketSenderWithDialer(customDialer)
	if sender == nil {
		t.Fatal("Expected non-nil sender")
	}
	if sender.dialFunc == nil {
		t.Fatal("Expected dialFunc to be set")
	}
}

// Helper mock types and functions

type mockPacketConn struct {
	closeFunc       func() error
	localAddrFunc   func() net.Addr
	readFromFunc    func(b []byte) (int, net.Addr, error)
	writeToFunc     func(b []byte, addr net.Addr) (int, error)
	setDeadlineFunc func(t time.Time) error
}

func (m *mockPacketConn) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockPacketConn) LocalAddr() net.Addr {
	if m.localAddrFunc != nil {
		return m.localAddrFunc()
	}
	return nil
}

func (m *mockPacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	if m.readFromFunc != nil {
		return m.readFromFunc(b)
	}
	return 0, nil, fmt.Errorf("not implemented")
}

func (m *mockPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if m.writeToFunc != nil {
		return m.writeToFunc(b, addr)
	}
	return 0, fmt.Errorf("not implemented")
}

func (m *mockPacketConn) SetDeadline(t time.Time) error {
	if m.setDeadlineFunc != nil {
		return m.setDeadlineFunc(t)
	}
	return nil
}

func (m *mockPacketConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (m *mockPacketConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
