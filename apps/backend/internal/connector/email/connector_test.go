package email

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/sp3640/opspilot/backend/internal/connector"
	"github.com/sp3640/opspilot/backend/internal/notification"
)

func TestEmailConnectorTestConnectionSucceedsForValidSMTPConfig(t *testing.T) {
	server := startTestSMTPServer(t, func(line string) string {
		switch {
		case strings.HasPrefix(line, "EHLO "), strings.HasPrefix(line, "HELO "):
			return "250-localhost\r\n250-STARTTLS\r\n250 AUTH PLAIN LOGIN"
		case strings.HasPrefix(line, "AUTH "):
			return "235 2.7.0 Authentication successful"
		case strings.HasPrefix(line, "MAIL FROM:"):
			return "250 2.1.0 OK"
		case strings.HasPrefix(line, "RCPT TO:"):
			return "250 2.1.5 OK"
		case line == "DATA":
			return "354 End data with <CR><LF>.<CR><LF>"
		case line == ".":
			return "250 2.0.0 OK: queued"
		default:
			return "250 OK"
		}
	})
	defer server.Close()

	c := NewConnector()
	result, err := c.TestConnection(context.Background(), connector.Config{
		Metadata:    map[string]string{"smtp_host": server.host, "smtp_port": server.port, "from": "ops@example.com", "to": "alerts@example.com"},
		Credentials: map[string]string{"username": "tester", "password": "hunter2"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success=true, got %+v", result)
	}
}

func TestEmailConnectorSendNotificationSucceeds(t *testing.T) {
	server := startTestSMTPServer(t, func(line string) string {
		switch {
		case strings.HasPrefix(line, "EHLO "), strings.HasPrefix(line, "HELO "):
			return "250-localhost\r\n250 AUTH PLAIN LOGIN"
		case strings.HasPrefix(line, "AUTH "):
			return "235 2.7.0 Authentication successful"
		case strings.HasPrefix(line, "MAIL FROM:"):
			return "250 2.1.0 OK"
		case strings.HasPrefix(line, "RCPT TO:"):
			return "250 2.1.5 OK"
		case line == "DATA":
			return "354 End data with <CR><LF>.<CR><LF>"
		case line == ".":
			return "250 2.0.0 OK: queued"
		default:
			return "250 OK"
		}
	})
	defer server.Close()

	c := NewConnector()
	msg := notification.Message{
		Title:    "HIGH Alert",
		Body:     "Pod crash detected",
		Severity: "HIGH",
		Fields:   map[string]string{"Application": "OpsPilot Backend", "Incident": "INC-123"},
	}
	if err := c.SendNotification(context.Background(), connector.Config{
		Metadata:    map[string]string{"smtp_host": server.host, "smtp_port": server.port, "from": "ops@example.com", "to": "alerts@example.com"},
		Credentials: map[string]string{"username": "tester", "password": "hunter2"},
	}, msg); err != nil {
		t.Fatalf("send notification: %v", err)
	}
	if server.lastFrom != "<ops@example.com>" && server.lastFrom != "ops@example.com" {
		t.Fatalf("expected sender to be recorded, got %q", server.lastFrom)
	}
	if server.lastTo != "alerts@example.com" {
		t.Fatalf("expected recipient to be recorded, got %q", server.lastTo)
	}
	if !strings.Contains(server.lastData, "Subject: HIGH Alert") || !strings.Contains(server.lastData, "Pod crash detected") || !strings.Contains(server.lastData, "Application: OpsPilot Backend") {
		t.Fatalf("expected operational context in SMTP payload, got %q", server.lastData)
	}
}

func TestEmailConnectorRejectsInvalidConfiguration(t *testing.T) {
	c := NewConnector()
	if _, err := c.TestConnection(context.Background(), connector.Config{}); err == nil {
		t.Fatal("expected empty config to fail")
	}
	if _, err := c.TestConnection(context.Background(), connector.Config{Metadata: map[string]string{"smtp_host": "smtp.example.com", "from": "ops@example.com"}}); err == nil {
		t.Fatal("expected config missing recipient to fail")
	}
}

func TestEmailConnectorDoesNotLeakPasswordInErrors(t *testing.T) {
	secret := "super-secret-smtp-password"
	server := startTestSMTPServer(t, func(line string) string {
		if strings.HasPrefix(line, "AUTH") {
			return "535 5.7.8 Authentication credentials invalid"
		}
		if strings.HasPrefix(line, "EHLO ") || strings.HasPrefix(line, "HELO ") {
			return "250-localhost\r\n250 AUTH PLAIN LOGIN"
		}
		return "250 OK"
	})
	defer server.Close()

	c := NewConnector()
	_, err := c.TestConnection(context.Background(), connector.Config{
		Metadata:    map[string]string{"smtp_host": server.host, "smtp_port": server.port, "from": "ops@example.com", "to": "alerts@example.com"},
		Credentials: map[string]string{"username": "tester", "password": secret},
	})
	if err == nil {
		t.Fatal("expected auth failure")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("expected password to stay out of errors, got %q", err)
	}
}

func TestEmailConnectorReportsCapabilities(t *testing.T) {
	caps := NewConnector().Capabilities()
	if !caps.Implemented || !caps.SupportsTestConnection || !caps.SupportsHealthCheck {
		t.Fatalf("expected full connector capabilities, got %+v", caps)
	}
}

type testSMTPServer struct {
	host     string
	port     string
	ln       net.Listener
	lastFrom string
	lastTo   string
	lastData string
}

func (s *testSMTPServer) Close() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func startTestSMTPServer(t *testing.T, handler func(string) string) *testSMTPServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := &testSMTPServer{host: "127.0.0.1", port: fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port), ln: ln}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleSMTPConn(conn, server, handler)
		}
	}()
	return server
}

func handleSMTPConn(conn net.Conn, server *testSMTPServer, handler func(string) string) {
	defer conn.Close()
	if _, err := conn.Write([]byte("220 localhost ESMTP\r\n")); err != nil {
		return
	}

	reader := bufio.NewReader(conn)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			return
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "MAIL FROM:") {
			server.lastFrom = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "MAIL FROM:")), "<>")
		}
		if strings.HasPrefix(line, "RCPT TO:") {
			server.lastTo = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "RCPT TO:")), "<>")
		}

		response := handler(line)
		if _, err := conn.Write([]byte(response + "\r\n")); err != nil {
			return
		}
		if strings.Contains(line, "QUIT") {
			return
		}
		if line == "DATA" {
			var data strings.Builder
			for {
				if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
					return
				}
				chunk, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				trimmed := strings.TrimSpace(chunk)
				if trimmed == "." {
					server.lastData = data.String()
					break
				}
				data.WriteString(trimmed)
				data.WriteString("\n")
			}
			if _, err := conn.Write([]byte("250 2.0.0 OK: queued\r\n")); err != nil {
				return
			}
		}
	}
}
