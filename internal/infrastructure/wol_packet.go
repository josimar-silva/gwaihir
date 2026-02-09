// Package infrastructure provides infrastructure layer implementations.
package infrastructure

import (
	"fmt"
	"net"
	"strings"
)

// WoLPacketSender handles sending Wake-on-LAN magic packets.
type WoLPacketSender struct {
	// For testing purposes, allows mocking the UDP connection
	dialFunc func(network, addr string) (net.PacketConn, error)
}

// NewWoLPacketSender creates a new WoL packet sender with default UDP dialer.
func NewWoLPacketSender() *WoLPacketSender {
	return &WoLPacketSender{
		dialFunc: net.ListenPacket,
	}
}

// NewWoLPacketSenderWithDialer creates a WoL packet sender with a custom dialer (for testing).
func NewWoLPacketSenderWithDialer(dialFunc func(network, addr string) (net.PacketConn, error)) *WoLPacketSender {
	return &WoLPacketSender{
		dialFunc: dialFunc,
	}
}

// SendMagicPacket sends a Wake-on-LAN magic packet to the specified MAC address on the broadcast address.
// The magic packet format: 6 bytes of 0xFF followed by 16 repetitions of the 6-byte MAC address.
func (s *WoLPacketSender) SendMagicPacket(mac, broadcast string) error {
	// Normalize MAC address to colon-separated format
	normalizedMAC := normalizeMACAddress(mac)

	// Parse MAC address bytes
	macBytes, err := net.ParseMAC(normalizedMAC)
	if err != nil {
		return fmt.Errorf("invalid MAC address format: %w", err)
	}

	// Create magic packet: 6 bytes of 0xFF + 16x MAC address
	packet := buildMagicPacket(macBytes)

	// Add port 9 (standard WoL port) to broadcast address
	broadcastAddr := net.JoinHostPort(broadcast, "9")

	// Listen on UDP to send the packet
	conn, err := s.dialFunc("udp", "0.0.0.0:0")
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Enable broadcast option on the connection
	if udpConn, ok := conn.(*net.UDPConn); ok {
		if err := udpConn.SetWriteBuffer(1024); err != nil {
			return fmt.Errorf("failed to set write buffer: %w", err)
		}
	}

	// Resolve broadcast address
	addr, err := net.ResolveUDPAddr("udp", broadcastAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve broadcast address '%s': %w", broadcastAddr, err)
	}

	// Send the magic packet
	if _, err := conn.WriteTo(packet, addr); err != nil {
		return fmt.Errorf("failed to send magic packet to %s: %w", broadcast, err)
	}

	return nil
}

// normalizeMACAddress converts MAC address to colon-separated format (uppercase).
// Supports both colon (AA:BB:CC:DD:EE:FF) and dash (AA-BB-CC-DD-EE-FF) formats.
func normalizeMACAddress(mac string) string {
	return strings.ToUpper(strings.ReplaceAll(mac, "-", ":"))
}

// buildMagicPacket constructs a Wake-on-LAN magic packet.
// Format: 6 bytes of 0xFF followed by the MAC address repeated 16 times.
func buildMagicPacket(macBytes net.HardwareAddr) []byte {
	packet := make([]byte, 102) // 6 + 16*6 = 102 bytes

	// Fill first 6 bytes with 0xFF
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}

	// Repeat MAC address 16 times starting at byte 6
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], macBytes)
	}

	return packet
}
