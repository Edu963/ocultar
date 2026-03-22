package audit

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Event struct {
	Timestamp         string `json:"timestamp"`
	User              string `json:"user"`
	Action            string `json:"action"`
	Result            string `json:"result"`
	ComplianceMapping string `json:"compliance_mapping,omitempty"` // e.g., "GDPR", "HIPAA", "Business Secrets"
}

type EnterpriseLogger struct {
	eventsCh chan Event
	logFile  *os.File
}

func NewLogger() *EnterpriseLogger {
	return &EnterpriseLogger{}
}

func (l *EnterpriseLogger) Init(filePath string) error {
	var err error
	l.logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.eventsCh = make(chan Event, 10000)

	go func() {
		encoder := json.NewEncoder(l.logFile)
		for event := range l.eventsCh {
			if err := encoder.Encode(event); err != nil {
				log.Printf("[ERR] Audit logger serialization failed: %v", err)
			}
		}
	}()

	log.Printf("[INFO] Enterprise Async Audit Logger initialized -> %s", filePath)
	return nil
}

func (l *EnterpriseLogger) Log(user, action, result, mapping string) {
	if l.eventsCh == nil {
		return
	}

	event := Event{
		Timestamp:         time.Now().UTC().Format(time.RFC3339Nano),
		User:              user,
		Action:            action,
		Result:            result,
		ComplianceMapping: mapping,
	}

	// Block with a timeout; if the log goroutine is overwhelmed (disk I/O stall),
	// halt the proxy rather than silently dropping a GDPR-required audit event.
	select {
	case l.eventsCh <- event:
	case <-time.After(2 * time.Second):
		log.Fatalf("[FATAL] Audit log buffer deadlocked. Halting proxy to protect GDPR Article 5(2) accountability. Check disk I/O on audit log path.")
	}
}

func (l *EnterpriseLogger) Close() {
	if l.eventsCh != nil {
		close(l.eventsCh)
	}
	if l.logFile != nil {
		l.logFile.Close()
	}
}
