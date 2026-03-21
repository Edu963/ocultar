package engine_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/Edu963/ocultar/pkg/config"
	"github.com/Edu963/ocultar/pkg/engine"
	"github.com/Edu963/ocultar/pkg/vault"
)

type qaTestCase struct {
	Name          string
	Input         string
	StrictPattern string // Regex that should yield 0 matches after refinement
}

func TestFalseNegativeKPI(t *testing.T) {
	_ = os.Remove("test_vault.db")
	v, err := vault.New(config.Settings{VaultBackend: "duckdb"}, "test_vault.db")
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	defer v.Close()
	defer os.Remove("test_vault.db")

	masterKey := []byte("01234567890123456789012345678901") // 32 bytes

	// 2. Define the PII corpus (edge cases, camouflage, international formats)
	corpus := []qaTestCase{
		{
			Name:          "Standard Email",
			Input:         "Contact me at jeff.bezos@amazon.com tomorrow.",
			StrictPattern: `(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}`,
		},
		{
			Name:          "French Mobile Phone",
			Input:         "Call my personal number: +33 6 12 34 56 78 please.",
			StrictPattern: `(?i)\+33[\s.-]*[67][\s.-]*(\d{2}[\s.-]*){3}\d{2}`,
		},
		{
			Name:          "North American Phone",
			Input:         "Support desk: 1-800-555-0199.",
			StrictPattern: `(?i)1-800-555-0199`,
		},
		{
			Name:          "Hidden Structure (CRLFs)",
			Input:         "My email is sneaky\r\njohn@sneaky.org\r\nand you won't find it.",
			StrictPattern: `(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}`,
		},
		{
			Name:          "European Address",
			Input:         "We shipped the hardware to 14 Boulevard MacDonald, Paris.",
			StrictPattern: `(?i)14 Boulevard MacDonald`, // Exact match should be gone
		},
		{
			Name:          "URL / Web Link",
			Input:         "Download the file from https://secure-bucket.s3.amazonaws.com/data.csv",
			StrictPattern: `(?i)https?://[a-zA-Z0-9.-]+`,
		},
		{
			Name:          "Timestamp Bug (Should not become phone)",
			Input:         "2026-02-21T10:00:00Z",
			StrictPattern: `\[PHONE_`, // Verify it does not generate a phone token
		},
		{
			Name:          "Overlap Bug (Nested tokens)",
			Input:         "Email david.smith@company.net with phone 12345678",
			StrictPattern: `\[EMAIL_\[PHONE_`, // Verify no nested tokens are generated
		},
	}

	// 3. Execute the pipeline and assert Zero Leaks
	config.InitDefaults()

	for _, tc := range corpus {
		t.Run(tc.Name, func(t *testing.T) {
			eng := engine.NewEngine(v, masterKey)
			eng.DryRun = false
			eng.Report = false

			refined, err := eng.RefineString(tc.Input, "qa-tester", nil)
			if err != nil {
				t.Fatalf("Engine failure: %v", err)
			}

			// The strict pattern should NOT be found in the transformed text
			re := regexp.MustCompile(tc.StrictPattern)
			if re.MatchString(refined) {
				t.Errorf("🚨 FALSE NEGATIVE DETECTED!\nOriginal: %s\nRefined:  %s\nLeaked:   %s",
					tc.Input, refined, re.FindString(refined))
			} else {
				t.Logf("✅ Cleaned %s -> %s", tc.Name, refined)
			}
		})
	}
}

// MockAIScanner tracks how many times ScanForPII is called.
type MockAIScanner struct {
	engine.NoopAIScanner
	ScanCount int
}

func (m *MockAIScanner) ScanForPII(text string) (map[string][]string, error) {
	m.ScanCount++
	// Return a hit to simulate PII detection
	return map[string][]string{
		"PERSON": {"John Doe"},
	}, nil
}

func (m *MockAIScanner) IsAvailable() bool { return true }

func TestSinglePassScanning(t *testing.T) {
	config.InitDefaults()
	masterKey := []byte("01234567890123456789012345678901")
	v, _ := vault.New(config.Settings{VaultBackend: "duckdb"}, ":memory:")
	defer v.Close()

	mock := &MockAIScanner{}
	eng := engine.NewEngine(v, masterKey)
	eng.AIScanner = mock

	// Complex nested structure with Base64 and nested JSON
	data := map[string]interface{}{
		"user": "John Doe",
		"details": map[string]interface{}{
			"bio": "My name is John Doe and I live in Paris.",
			"nested_json": `{"manager": "John Doe"}`,
		},
	}

	_, err := eng.ProcessInterface(data, "test-actor")
	if err != nil {
		t.Fatalf("ProcessInterface failed: %v", err)
	}

	if mock.ScanCount != 1 {
		t.Errorf("🚨 REDUNDANT SCANS DETECTED! ScanForPII called %d times, expected 1", mock.ScanCount)
	} else {
		t.Logf("✅ Single-pass scan verified: ScanForPII called exactly %d time", mock.ScanCount)
	}
}
