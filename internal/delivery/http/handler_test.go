package http //nolint:revive

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/josimar-silva/gwaihir/internal/domain"
	"github.com/josimar-silva/gwaihir/internal/usecase"
)

// Mock repository for testing
type mockRepository struct {
	machines map[string]*domain.Machine
}

func (m *mockRepository) GetByID(id string) (*domain.Machine, error) {
	machine, ok := m.machines[id]
	if !ok {
		return nil, domain.ErrMachineNotFound
	}
	return machine, nil
}

func (m *mockRepository) GetAll() ([]*domain.Machine, error) {
	machines := make([]*domain.Machine, 0, len(m.machines))
	for _, machine := range m.machines {
		machines = append(machines, machine)
	}
	return machines, nil
}

func (m *mockRepository) Exists(id string) bool {
	_, ok := m.machines[id]
	return ok
}

// Mock WoL packet sender for testing
type mockPacketSender struct {
	callCount       int
	lastMac         string
	lastBroadcast   string
	sendError       error
	shouldFailCount int
}

func (m *mockPacketSender) SendMagicPacket(mac, broadcast string) error {
	m.callCount++
	m.lastMac = mac
	m.lastBroadcast = broadcast

	if m.shouldFailCount > 0 {
		m.shouldFailCount--
		return m.sendError
	}
	return nil
}

// Test helper to create a handler with mocks
func newHandlerForTesting(machines map[string]*domain.Machine) (*Handler, *mockRepository, *mockPacketSender) {
	if machines == nil {
		machines = map[string]*domain.Machine{
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
	}

	repo := &mockRepository{machines: machines}
	sender := &mockPacketSender{}
	wolUseCase := usecase.NewWoLUseCase(repo, sender)
	handler := NewHandler(wolUseCase, "0.1.0", "2024-01-01T00:00:00Z", "abc123")

	return handler, repo, sender
}

// Tests

func TestNewHandler(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
	if handler.version != "0.1.0" {
		t.Errorf("Expected version 0.1.0, got %s", handler.version)
	}
	if handler.buildTime != "2024-01-01T00:00:00Z" {
		t.Errorf("Expected buildTime, got %s", handler.buildTime)
	}
	if handler.gitCommit != "abc123" {
		t.Errorf("Expected gitCommit, got %s", handler.gitCommit)
	}
}

func TestHandler_Response_Types(t *testing.T) {
	// Test response type constructors work
	successResp := SuccessResponse{Message: "test"}
	if successResp.Message != "test" {
		t.Errorf("Expected SuccessResponse.Message to be 'test', got %s", successResp.Message)
	}

	errorResp := ErrorResponse{Error: "error"}
	if errorResp.Error != "error" {
		t.Errorf("Expected ErrorResponse.Error to be 'error', got %s", errorResp.Error)
	}

	versionResp := VersionResponse{
		Version:   "1.0.0",
		BuildTime: "2024-01-01",
		GitCommit: "abc123",
	}
	if versionResp.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", versionResp.Version)
	}
}

const testMachineSaruman = "saruman"

func TestHandler_WakeRequest(t *testing.T) {
	req := WakeRequest{MachineID: testMachineSaruman}
	if req.MachineID != testMachineSaruman {
		t.Errorf("Expected MachineID saruman, got %s", req.MachineID)
	}
}

func TestHandler_Integration(t *testing.T) {
	handler, repo, sender := newHandlerForTesting(nil)

	// Verify handler was created with correct usecase
	if handler.wolUseCase == nil {
		t.Fatal("Expected wolUseCase to be set")
	}

	// Verify repository works
	machine, err := repo.GetByID("saruman")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if machine == nil {
		t.Fatal("Expected machine, got nil")
	}
	if machine.ID != "saruman" {
		t.Errorf("Expected ID saruman, got %s", machine.ID)
	}

	// Verify sender tracks calls
	if sender.callCount != 0 {
		t.Errorf("Expected no calls initially, got %d", sender.callCount)
	}
}

func TestHandler_EmptyMachineList(t *testing.T) {
	handler, _, _ := newHandlerForTesting(map[string]*domain.Machine{})

	machines, err := handler.wolUseCase.ListMachines()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(machines) != 0 {
		t.Errorf("Expected empty list, got %d machines", len(machines))
	}
}

func TestHandler_MultipleMachines(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)

	machines, err := handler.wolUseCase.ListMachines()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(machines) != 2 {
		t.Errorf("Expected 2 machines, got %d", len(machines))
	}
}

// HTTP Endpoint Tests

func TestHTTP_Wake_Success(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouter(handler)

	reqBody := WakeRequest{MachineID: "saruman"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/wol", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	var resp SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
	if resp.Message != "WoL packet sent successfully" {
		t.Errorf("Expected success message, got %s", resp.Message)
	}
}

func TestHTTP_Wake_MachineNotFound(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouter(handler)

	reqBody := WakeRequest{MachineID: "nonexistent"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/wol", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
	if resp.Error != "Machine not found or not allowed" {
		t.Errorf("Expected not found error, got %s", resp.Error)
	}
}

func TestHTTP_Wake_InvalidJSON(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/wol", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHTTP_Wake_MissingMachineID(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouter(handler)

	reqBody := WakeRequest{} // Empty machine_id
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/wol", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHTTP_ListMachines(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/machines", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var machines []*domain.Machine
	err := json.Unmarshal(w.Body.Bytes(), &machines)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
	if len(machines) != 2 {
		t.Errorf("Expected 2 machines, got %d", len(machines))
	}
}

func TestHTTP_ListMachines_Empty(t *testing.T) {
	handler, _, _ := newHandlerForTesting(map[string]*domain.Machine{})
	router := NewRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/machines", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var machines []*domain.Machine
	if err := json.Unmarshal(w.Body.Bytes(), &machines); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
	if machines == nil || len(machines) != 0 {
		t.Errorf("Expected empty list, got %v", machines)
	}
}

func TestHTTP_GetMachine(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/machines/saruman", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var machine domain.Machine
	err := json.Unmarshal(w.Body.Bytes(), &machine)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
	if machine.ID != "saruman" {
		t.Errorf("Expected ID saruman, got %s", machine.ID)
	}
	if machine.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("Expected MAC AA:BB:CC:DD:EE:FF, got %s", machine.MAC)
	}
}

func TestHTTP_GetMachine_NotFound(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/machines/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
	if resp.Error != "Machine not found" {
		t.Errorf("Expected 'Machine not found', got %s", resp.Error)
	}
}

func TestHTTP_Health(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp gin.H
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
	if resp["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", resp["status"])
	}
}

func TestHTTP_Version(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp VersionResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}
	if resp.Version != "0.1.0" {
		t.Errorf("Expected version 0.1.0, got %s", resp.Version)
	}
	if resp.BuildTime != "2024-01-01T00:00:00Z" {
		t.Errorf("Expected buildTime 2024-01-01T00:00:00Z, got %s", resp.BuildTime)
	}
	if resp.GitCommit != "abc123" {
		t.Errorf("Expected gitCommit abc123, got %s", resp.GitCommit)
	}
}

func TestHTTP_WakeWithPacketError(t *testing.T) {
	handler, _, sender := newHandlerForTesting(nil)
	router := NewRouter(handler)

	// Make sender fail
	sender.shouldFailCount = 1
	sender.sendError = &gin.Error{Err: nil, Type: gin.ErrorTypeBind, Meta: "network error"}

	reqBody := WakeRequest{MachineID: "saruman"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/wol", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
