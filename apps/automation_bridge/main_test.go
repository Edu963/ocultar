package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// chdirTemp changes the working directory to dir and restores it after the test.
func chdirTemp(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

// resetGlobals snapshots and restores the package globals around each test.
func resetGlobals(t *testing.T) {
	t.Helper()
	origConfig := config
	historyMu.Lock()
	origHistory := make([]RunHistory, len(historyList))
	copy(origHistory, historyList)
	historyMu.Unlock()

	t.Cleanup(func() {
		historyMu.Lock()
		config = origConfig
		historyList = origHistory
		historyMu.Unlock()
	})
}

// ── loadConfig ────────────────────────────────────────────────────────────────

func TestLoadConfig_Success(t *testing.T) {
	resetGlobals(t)
	// Default cwd during `go test` is the package dir (apps/automation_bridge).
	// ../../configs/automation_commands.json resolves to the repo root configs/.
	if err := loadConfig(); err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(config.Commands) == 0 {
		t.Error("expected at least one command in loaded config")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	resetGlobals(t)
	// Chdir to an empty tmpdir — no configs/ subdirectory exists.
	chdirTemp(t, t.TempDir())
	if err := loadConfig(); err == nil {
		t.Error("expected error for missing config file, got nil")
	}
}

func TestLoadConfig_MalformedJSON(t *testing.T) {
	resetGlobals(t)
	tmp := t.TempDir()
	// Recreate the relative path ../../configs/automation_commands.json
	// from inside a fake apps/automation_bridge directory.
	fakePkg := filepath.Join(tmp, "apps", "automation_bridge")
	fakeConfigs := filepath.Join(tmp, "configs")
	_ = os.MkdirAll(fakePkg, 0o755)
	_ = os.MkdirAll(fakeConfigs, 0o755)
	_ = os.WriteFile(filepath.Join(fakeConfigs, "automation_commands.json"), []byte("{bad json!!"), 0o644)
	chdirTemp(t, fakePkg)
	if err := loadConfig(); err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

// ── historyHandler ────────────────────────────────────────────────────────────

func TestHistoryHandler_Empty(t *testing.T) {
	resetGlobals(t)
	historyMu.Lock()
	historyList = nil
	historyMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/automation/history", nil)
	rr := httptest.NewRecorder()
	historyHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	// json.Encoder encodes nil slice as "null"; unmarshal into []RunHistory is nil, not error.
	body := strings.TrimSpace(rr.Body.String())
	if body != "null" && body != "[]" {
		t.Errorf("expected null or [] for empty history, got %q", body)
	}
}

func TestHistoryHandler_WithEntries(t *testing.T) {
	resetGlobals(t)
	now := time.Now()
	historyMu.Lock()
	historyList = []RunHistory{
		{RunID: "run_1", CommandID: "run_security_scan", Status: "success", StartTime: now},
		{RunID: "run_2", CommandID: "provision_llm", Status: "running", StartTime: now},
	}
	historyMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/automation/history", nil)
	rr := httptest.NewRecorder()
	historyHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got []RunHistory
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 entries, got %d", len(got))
	}
	if got[0].RunID != "run_1" || got[1].RunID != "run_2" {
		t.Errorf("unexpected run IDs: %v, %v", got[0].RunID, got[1].RunID)
	}
}

// ── updateHistory ─────────────────────────────────────────────────────────────

func TestUpdateHistory_ThreadSafety(t *testing.T) {
	resetGlobals(t)
	ids := []string{"run_a", "run_b", "run_c", "run_d", "run_e"}
	historyMu.Lock()
	historyList = make([]RunHistory, len(ids))
	for i, id := range ids {
		historyList[i] = RunHistory{RunID: id, Status: "running"}
	}
	historyMu.Unlock()

	var wg sync.WaitGroup
	for _, id := range ids {
		id := id
		wg.Add(1)
		go func() {
			defer wg.Done()
			updateHistory(id, "success", "concurrent log")
		}()
	}
	wg.Wait()

	historyMu.Lock()
	defer historyMu.Unlock()
	for _, h := range historyList {
		if h.Status != "success" {
			t.Errorf("run %s still has status %q, expected success", h.RunID, h.Status)
		}
	}
}

// ── runHandler ────────────────────────────────────────────────────────────────

func TestRunHandler_InvalidJSON(t *testing.T) {
	resetGlobals(t)
	req := httptest.NewRequest(http.MethodPost, "/api/automation/run", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	runHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON body, got %d", rr.Code)
	}
}

func TestRunHandler_UnknownCommand(t *testing.T) {
	resetGlobals(t)
	if err := loadConfig(); err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	body := `{"command_id":"absolutely_nonexistent_xyz","inputs":{}}`
	req := httptest.NewRequest(http.MethodPost, "/api/automation/run", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	runHandler(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown command, got %d", rr.Code)
	}
}

func TestRunHandler_MissingRequiredInput(t *testing.T) {
	resetGlobals(t)
	// Inject a minimal command with one required input directly into the config.
	config.Commands = map[string]CommandDef{
		"need_key": {
			ID:      "need_key",
			Name:    "Requires Key",
			Command: "echo",
			Args:    []string{"hello"},
			Inputs: []InputDef{
				{Name: "OCU_MASTER_KEY", Type: "text", Required: true, Description: "master key"},
			},
		},
	}
	body := `{"command_id":"need_key","inputs":{}}`
	req := httptest.NewRequest(http.MethodPost, "/api/automation/run", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	runHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing required input, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "OCU_MASTER_KEY") {
		t.Errorf("expected error to mention missing field name, got: %s", rr.Body.String())
	}
}

// ── commandsHandler ───────────────────────────────────────────────────────────

func TestCommandsHandler(t *testing.T) {
	resetGlobals(t)
	if err := loadConfig(); err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/automation/commands", nil)
	rr := httptest.NewRecorder()
	commandsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got []CommandDef
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode commands: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected at least one command in response")
	}
	for _, cmd := range got {
		if cmd.ID == "" {
			t.Errorf("command missing ID: %+v", cmd)
		}
		if cmd.Command == "" {
			t.Errorf("command %q has no executable set", cmd.ID)
		}
	}
}

// ── corsMiddleware ─────────────────────────────────────────────────────────────

func TestCORSMiddleware_SetsHeaders(t *testing.T) {
	handler := corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected CORS origin *, got %q", got)
	}
}

func TestCORSMiddleware_OPTIONS(t *testing.T) {
	handler := corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Should not be reached for OPTIONS
		t.Error("inner handler called for OPTIONS preflight")
	})
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for OPTIONS, got %d", rr.Code)
	}
}
