package notification

import (
	"bufio"
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeSMTPServer speaks just enough real SMTP for net/smtp to talk to, and -
// like real servers - insists the envelope sender is a bare address.
type fakeSMTPServer struct {
	addr string

	mu       sync.Mutex
	mailFrom string
	rcptTo   []string
	data     string
	rejected string
}

func startFakeSMTPServer(t *testing.T) *fakeSMTPServer {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := &fakeSMTPServer{addr: listener.Addr().String()}
	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.serve(conn)
		}
	}()
	return server
}

func (s *fakeSMTPServer) serve(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(conn)
	reply := func(line string) { _, _ = conn.Write([]byte(line + "\r\n")) }

	reply("220 fake ESMTP")
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		upper := strings.ToUpper(line)
		switch {
		case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
			reply("250 fake")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			from := strings.TrimSpace(line[len("MAIL FROM:"):])
			if !strings.HasPrefix(from, "<") || strings.ContainsAny(from, `" `) {
				s.mu.Lock()
				s.rejected = from
				s.mu.Unlock()
				reply(`501 5.5.4 Syntax error in parameters or arguments (invalid FROM parameter)`)
				continue
			}
			s.mu.Lock()
			s.mailFrom = strings.Trim(from, "<>")
			s.mu.Unlock()
			reply("250 ok")
		case strings.HasPrefix(upper, "RCPT TO:"):
			s.mu.Lock()
			s.rcptTo = append(s.rcptTo, strings.Trim(strings.TrimSpace(line[len("RCPT TO:"):]), "<>"))
			s.mu.Unlock()
			reply("250 ok")
		case upper == "DATA":
			reply("354 go ahead")
			var body strings.Builder
			for {
				part, err := reader.ReadString('\n')
				if err != nil || part == ".\r\n" {
					break
				}
				body.WriteString(part)
			}
			s.mu.Lock()
			s.data = body.String()
			s.mu.Unlock()
			reply("250 queued")
		case upper == "QUIT":
			reply("221 bye")
			return
		default:
			reply("250 ok")
		}
	}
}

func (s *fakeSMTPServer) host() (string, string) {
	host, port, _ := net.SplitHostPort(s.addr)
	return host, port
}

func TestEmailProviderSendsOverRealSMTPWithADisplayNameSender(t *testing.T) {
	server := startFakeSMTPServer(t)
	host, port := server.host()
	provider := NewEmailProvider(EmailConfig{Host: host, Port: port, From: "OpsPilot <no-reply@mail.example.com>"})

	err := provider.Send(context.Background(), "user@example.com", Message{Title: "Verify your OpsPilot email", Body: "hello", HTML: "<p>hello</p>"})
	if err != nil {
		t.Fatalf("send over SMTP: %v (a display name in SMTP_FROM must not reach MAIL FROM)", err)
	}

	server.mu.Lock()
	defer server.mu.Unlock()
	if server.mailFrom != "no-reply@mail.example.com" {
		t.Fatalf("envelope sender = %q, want the bare address", server.mailFrom)
	}
	if len(server.rcptTo) != 1 || server.rcptTo[0] != "user@example.com" {
		t.Fatalf("recipients = %v", server.rcptTo)
	}
	if !strings.Contains(server.data, `From: "OpsPilot" <no-reply@mail.example.com>`) {
		t.Fatalf("the From header lost the display name: %q", server.data)
	}
	if !strings.Contains(server.data, "Subject: Verify your OpsPilot email") || !strings.Contains(server.data, "multipart/alternative") {
		t.Fatalf("unexpected message: %q", server.data)
	}
}

func TestEmailProviderSendsOverRealSMTPWithABareSender(t *testing.T) {
	server := startFakeSMTPServer(t)
	host, port := server.host()
	provider := NewEmailProvider(EmailConfig{Host: host, Port: port, From: "no-reply@mail.example.com"})

	if err := provider.Send(context.Background(), "user@example.com", Message{Title: "t", Body: "b"}); err != nil {
		t.Fatalf("send over SMTP: %v", err)
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.mailFrom != "no-reply@mail.example.com" {
		t.Fatalf("envelope sender = %q", server.mailFrom)
	}
}
