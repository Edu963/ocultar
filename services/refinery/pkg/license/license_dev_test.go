package license

import (
	"os"
	"testing"
)

func TestGeneratedLicense(t *testing.T) {
	// The license we just generated
	licenseKey := "5Kbs4nRFC5wWOvAwp/tEmr66WIAO3f37wna47b7wCpnAhOL/bXt61jBWtp0XC/ARdRd7qjX35vazK6IwONZLBQ==.eyJDdXN0b21lck5hbWUiOiJEZXZlbG9wbWVudCIsIlRpZXIiOiJlbnRlcnByaXNlIiwiRXhwaXJ5RGF0ZSI6MTgwNTU4MDQ0MiwiQ2FwYWJpbGl0aWVzIjozfQ=="
	
	os.Setenv("OCU_LICENSE_KEY", licenseKey)
	defer os.Unsetenv("OCU_LICENSE_KEY")

	Load()

	if Active.Tier != "enterprise" {
		t.Errorf("Expected tier to be enterprise, got %s", Active.Tier)
	}

	if Active.CustomerName != "Development" {
		t.Errorf("Expected customer to be Development, got %s", Active.CustomerName)
	}

	if !HasProConnector(CapProSlack) {
		t.Errorf("Expected Slack capability to be enabled")
	}

	if !HasProConnector(CapProSharePoint) {
		t.Errorf("Expected SharePoint capability to be enabled")
	}
}
