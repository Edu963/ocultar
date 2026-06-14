package license

import "testing"

// Verify the OSS stub invariants: all features enabled, no license key required.
func TestLicenseStub(t *testing.T) {
	Load() // no-op — must not panic

	if !IsEnterprise() {
		t.Error("IsEnterprise() must always return true in the OSS build")
	}

	for _, cap := range []uint64{CapProSlack, CapProSharePoint} {
		if !HasProConnector(cap) {
			t.Errorf("HasProConnector(%d) must always return true in the OSS build", cap)
		}
	}
}
