package proxy

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/Edu963/ocultar/pkg/refinery"
)

// SyslogServer listens for incoming UDP syslog messages, redacts PII,
// and forwards them to an upstream SIEM.
type SyslogServer struct {
	eng          *refinery.Refinery
	listenAddr   string
	upstreamAddr string
	conn         *net.UDPConn
}

// NewSyslogServer creates a new Syslog proxy server
func NewSyslogServer(eng *refinery.Refinery, listenAddr, upstreamAddr string) *SyslogServer {
	return &SyslogServer{
		eng:          eng,
		listenAddr:   listenAddr,
		upstreamAddr: upstreamAddr,
	}
}

// Start opens the UDP listener and starts processing messages
func (s *SyslogServer) Start() error {
	addr, err := net.ResolveUDPAddr("udp", s.listenAddr)
	if err != nil {
		return err
	}
	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	var upstream *net.UDPAddr
	if s.upstreamAddr != "" {
		upstream, err = net.ResolveUDPAddr("udp", s.upstreamAddr)
		if err != nil {
			return fmt.Errorf("invalid upstream syslog addr: %w", err)
		}
	}

	go s.listen(upstream)
	return nil
}

func (s *SyslogServer) listen(upstream *net.UDPAddr) {
	buf := make([]byte, 65535)
	for {
		n, _, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			log.Printf("[SYSLOG] read error: %v", err)
			continue
		}

		msg := string(buf[:n])
		// Send through refinery for redaction
		refined, err := s.eng.RefineString(msg, "syslog_proxy", nil)
		if err != nil {
			log.Printf("[SYSLOG-BLOCK] Refinery error: %v (dropping message)", err)
			continue // Fail-closed: drop on error
		}

		if upstream != nil {
			// Forward to SIEM
			upstreamConn, err := net.DialUDP("udp", nil, upstream)
			if err == nil {
				upstreamConn.Write([]byte(refined))
				upstreamConn.Close()
			} else {
				log.Printf("[SYSLOG] Failed to dial upstream: %v", err)
			}
		} else {
			// Log locally if no upstream SIEM is configured
			log.Printf("[SYSLOG-REDACTED] %s", refined)
		}
	}
}

// Stop closes the UDP connection
func (s *SyslogServer) Stop() {
	if s.conn != nil {
		s.conn.Close()
	}
}
