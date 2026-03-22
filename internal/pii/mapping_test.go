package pii_test

import (
	"encoding/json"
	"testing"
	"github.com/Edu963/ocultar/internal/pii"
)

func TestMappingAlias(t *testing.T) {
	eng := pii.NewEngine()
	mapping := map[string]string{
		"EMAIL": "EMAIL_ADDRESS",
		"AWS_KEY": "AWS_CREDENTIALS",
	}
	eng.SetMapping(mapping)

	input := "My email is test@example.com and key is AKIA1234567890123456"
	res := eng.Scan(input)

	if len(res) < 2 {
		t.Fatalf("Expected at least 2 hits, got %d", len(res))
	}

	for _, hit := range res {
		if hit.Entity == "EMAIL" && hit.CanonicalType != "EMAIL_ADDRESS" {
			t.Errorf("Expected CanonicalType EMAIL_ADDRESS for EMAIL, got %q", hit.CanonicalType)
		}
		if hit.Entity == "AWS_KEY" && hit.CanonicalType != "AWS_CREDENTIALS" {
			t.Errorf("Expected CanonicalType AWS_CREDENTIALS for AWS_KEY, got %q", hit.CanonicalType)
		}
	}

	// Verify JSON serialization
	b, err := json.Marshal(res[0])
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}
	var raw map[string]interface{}
	json.Unmarshal(b, &raw)
	if _, ok := raw["canonical_type"]; !ok {
		t.Errorf("JSON output missing canonical_type field")
	}
}

func TestCloudSecrets(t *testing.T) {
	eng := pii.NewEngine()
	
	cases := []struct {
		name      string
		input      string
		expectType string
	}{
		{"AWS Key", "AKIA1234567890123456", "AWS_KEY"},
		{"AWS Secret", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "AWS_SECRET"},
		{"GCP Service Account", "my-account@my-project.iam.gserviceaccount.com", "GCP_SERVICE_ACCOUNT"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := eng.Scan(tc.input)
			if len(res) == 0 {
				t.Fatalf("Expected hit for %s, got none", tc.name)
			}
			if res[0].Entity != tc.expectType {
				t.Errorf("Expected type %s, got %s", tc.expectType, res[0].Entity)
			}
		})
	}
}

func TestIPAddress(t *testing.T) {
	eng := pii.NewEngine()
	input := "The server is at 192.168.1.1 and 10.0.0.255"
	res := eng.Scan(input)

	if len(res) != 2 {
		t.Fatalf("Expected 2 IP hits, got %d", len(res))
	}
	for _, hit := range res {
		if hit.Entity != "IP_ADDRESS" {
			t.Errorf("Expected IP_ADDRESS, got %s", hit.Entity)
		}
	}
}

func TestNordicIDs(t *testing.T) {
	eng := pii.NewEngine()
	
	cases := []struct {
		name       string
		input      string
		expectType string
		isValid    bool
	}{
		{"Valid SE PIN", "SE-PIN 19121212-1212 is here", "SE_PIN", true},
		{"Invalid SE PIN", "19121212-1213", "SE_PIN", false},
		{"Valid DK CPR", "111111-1118", "DK_CPR", true}, 
		{"Invalid DK CPR", "111111-1111", "DK_CPR", false},
        // FI HETU valid sample: 010101-001R
        {"Valid FI HETU", "010101-001R", "FI_HETU", true},
		{"Invalid FI HETU", "010101-001A", "FI_HETU", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := eng.Scan(tc.input)
			t.Logf("Scan %s (%q): results: %+v", tc.name, tc.input, res)
			if tc.isValid {
				if len(res) == 0 {
					t.Errorf("Expected valid %s to pass, got no hits", tc.name)
				} else if res[0].Entity != tc.expectType {
                    t.Errorf("Expected type %s, got %s", tc.expectType, res[0].Entity)
                }
			} else {
				if len(res) > 0 && res[0].Entity == tc.expectType {
					t.Errorf("Expected invalid %s to fail, but got hit", tc.name)
				}
			}
		})
	}
}
