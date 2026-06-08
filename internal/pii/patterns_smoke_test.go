package pii

import (
	"testing"
)

func TestNewPatterns(t *testing.T) {
	tests := []struct {
		id   string
		text string
	}{
		{"HOME_ADDRESS",         "Please process the following: 742 Evergreen Terrace, Springfield"},
		{"US_VOTER_ID",          "Please process the following: Voter ID: VR-123456"},
		{"DE_PERSONALAUSWEIS",   "Please process the following: L01X00T47"},
		{"PROFESSIONAL_LICENSE", "Please process the following: Bar #: CA-123456"},
		{"CHILD_DEVICE_ID",      "Please process the following: kid IDFA: A3C4E567-89AB-CDEF-0123-456789ABCDEF"},
	}
	r := NewRefinery()
	for _, tc := range tests {
		t.Run(tc.id, func(t *testing.T) {
			results := r.Scan(tc.text)
			for _, res := range results {
				if res.Entity == tc.id {
					return // matched
				}
			}
			t.Errorf("pattern %s did not match %q — got %d results: %v", tc.id, tc.text, len(results), results)
		})
	}
}
