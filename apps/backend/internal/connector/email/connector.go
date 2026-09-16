package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/sp3640/opspilot/backend/internal/connector"
	"github.com/sp3640/opspilot/backend/internal/notification"
)

const (
	testTimeout = 5 * time.Second
)

var (
	ErrSMTPHostRequired = errors.New("smtp host is required")
	ErrSMTPFromRequired = errors.New("smtp from address is required")
	ErrSMTPToRequired   = errors.New("smtp recipient is required")
)

type Connector struct{}

func NewConnector() *Connector { return &Connector{} }

func (c *Connector) Connect(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	return c.TestConnection(ctx, cfg)
}

func (c *Connector) Disconnect(_ context.Context, _ connector.Config) error { return nil }

func (c *Connector) TestConnection(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	params, err := parseConfig(cfg)
	if err != nil {
		return nil, err
	}

	if err := validateSMTPConfig(params); err != nil {
		return nil, err
	}

	msg := notification.Message{
		Title: "OpsPilot connectivity test",
		Body:  "SMTP connection test succeeded.",
	}
	if err := smtpDialWithContext(ctx, params, msg); err != nil {
		return nil, err
	}

	return &connector.Result{Success: true, Message: "SMTP connection test succeeded."}, nil
}

func (c *Connector) HealthCheck(ctx context.Context, cfg connector.Config) (*connector.Result, error) {
	return c.TestConnection(ctx, cfg)
}

func (c *Connector) Capabilities() connector.Capabilities {
	return connector.Capabilities{
		Implemented:            true,
		SupportsTestConnection: true,
		SupportsHealthCheck:    true,
		Description:            "Send operational alerts via SMTP using a configured sender and recipients.",
	}
}

func (c *Connector) SendNotification(ctx context.Context, cfg connector.Config, msg notification.Message) error {
	params, err := parseConfig(cfg)
	if err != nil {
		return err
	}
	if err := validateSMTPConfig(params); err != nil {
		return err
	}
	if err := smtpDialWithContext(ctx, params, msg); err != nil {
		return err
	}
	return nil
}

type smtpConfig struct {
	host     string
	port     int
	from     string
	to       string
	username string
	password string
	security string
}

func parseConfig(cfg connector.Config) (smtpConfig, error) {
	metadata := cfg.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}

	security := strings.TrimSpace(metadata["security"])
	if security == "" {
		security = "none"
	}

	portStr := strings.TrimSpace(metadata["smtp_port"])
	if portStr == "" {
		portStr = metadata["port"]
	}
	if portStr == "" {
		portStr = "587"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return smtpConfig{}, fmt.Errorf("invalid SMTP port %q", portStr)
	}

	host := strings.TrimSpace(metadata["smtp_host"])
	if host == "" {
		host = strings.TrimSpace(metadata["host"])
	}
	if host == "" {
		return smtpConfig{}, ErrSMTPHostRequired
	}

	from := strings.TrimSpace(metadata["from"])
	if from == "" {
		return smtpConfig{}, ErrSMTPFromRequired
	}

	to := strings.TrimSpace(metadata["to"])
	if to == "" {
		to = strings.TrimSpace(metadata["recipient"])
	}
	if to == "" {
		return smtpConfig{}, ErrSMTPToRequired
	}

	username := strings.TrimSpace(cfg.Credentials["username"])
	password := strings.TrimSpace(cfg.Credentials["password"])

	return smtpConfig{host: host, port: port, from: from, to: to, username: username, password: password, security: security}, nil
}

func validateSMTPConfig(cfg smtpConfig) error {
	if cfg.host == "" {
		return ErrSMTPHostRequired
	}
	if cfg.from == "" {
		return ErrSMTPFromRequired
	}
	if cfg.to == "" {
		return ErrSMTPToRequired
	}
	return nil
}

func smtpDialWithContext(ctx context.Context, cfg smtpConfig, msg notification.Message) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, testTimeout)
	defer cancel()

	dialer := &net.Dialer{Timeout: testTimeout}
	addr := net.JoinHostPort(cfg.host, strconv.Itoa(cfg.port))
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("connect to SMTP host %s: %w", cfg.host, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.host)
	if err != nil {
		return fmt.Errorf("initialize SMTP client: %w", err)
	}
	defer client.Close()

	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("SMTP HELO: %w", err)
	}

	if cfg.username != "" || cfg.password != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(smtp.PlainAuth("", cfg.username, cfg.password, cfg.host)); err != nil {
				return fmt.Errorf("SMTP authentication: %w", err)
			}
		} else {
			return fmt.Errorf("SMTP server does not support AUTH")
		}
	}

	if strings.EqualFold(cfg.security, "tls") || strings.EqualFold(cfg.security, "starttls") {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: cfg.host}); err != nil {
				return fmt.Errorf("SMTP STARTTLS: %w", err)
			}
		} else if strings.EqualFold(cfg.security, "starttls") {
			return fmt.Errorf("SMTP server does not support STARTTLS")
		}
	}

	if err := client.Mail(cfg.from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}
	if err := client.Rcpt(cfg.to); err != nil {
		return fmt.Errorf("SMTP RCPT TO: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}
	if _, err := writer.Write(renderSMTPMessage(cfg.from, cfg.to, msg)); err != nil {
		return fmt.Errorf("SMTP DATA write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("SMTP DATA close: %w", err)
	}

	return nil
}

func renderSMTPMessage(from, to string, msg notification.Message) []byte {
	var body bytes.Buffer
	body.WriteString("From: " + from + "\r\n")
	body.WriteString("To: " + to + "\r\n")
	body.WriteString("Subject: " + msg.Title + "\r\n")
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	body.WriteString("\r\n")
	if msg.Body != "" {
		body.WriteString(msg.Body)
	}
	for _, key := range sortedSMTPFields(msg.Fields) {
		body.WriteString("\n")
		body.WriteString(key)
	}
	body.WriteString("\r\n")
	return body.Bytes()
}

func sortedSMTPFields(fields map[string]string) []string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, k+": "+fields[k])
	}
	return lines
}
