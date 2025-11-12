package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/cloud-file-storage/internal/domain/value_objects"
)

type SMTPMailSender struct {
	host     string
	port     int
	name     string
	email    string
	password string
}

func NewSMTPMailSender(host string, port int, name string, email string, password string) *SMTPMailSender {
	return &SMTPMailSender{
		host:     host,
		port:     port,
		name:     name,
		email:    email,
		password: password,
	}
}

func (s *SMTPMailSender) SendMagicLink(ctx context.Context, tokenHash value_objects.TokenHash, to string, baseURL string) error {
	c, conn, err := s.Auth(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := c.Quit(); err != nil {
			_ = c.Close()
		}
		_ = conn.Close()
	}()

	link := fmt.Sprintf("%s/magic?token=%s", strings.TrimRight(baseURL, "/"), tokenHash.String())
	from := s.email

	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	h := make(textproto.MIMEHeader)
	h.Set("From", fmt.Sprintf("%s <%s>", s.name, from))
	h.Set("To", to)
	h.Set("Subject", "Your magic link")
	h.Set("MIME-Version", "1.0")
	h.Set("Content-Type", "text/plain; charset=UTF-8")
	h.Set("Content-Transfer-Encoding", "8bit")

	var sb strings.Builder
	for k, v := range h {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(strings.Join(v, ", "))
		sb.WriteString("\r\n")
	}
	sb.WriteString("\r\n")
	sb.WriteString("Use the link to sign in:\r\n")
	sb.WriteString(link)
	sb.WriteString("\r\n")

	if _, err := w.Write([]byte(sb.String())); err != nil {
		return err
	}
	return nil
}

func (s *SMTPMailSender) Auth(ctx context.Context) (*smtp.Client, net.Conn, error) {
	d := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	c, err := smtp.NewClient(conn, s.host)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	if err = c.Hello(s.host); err != nil {
		c.Close()
		conn.Close()
		return nil, nil, err
	}

	if ok, _ := c.Extension("STARTTLS"); ok {
		cfg := &tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12}
		if err = c.StartTLS(cfg); err != nil {
			c.Close()
			conn.Close()
			return nil, nil, err
		}
	}

	if ok, _ := c.Extension("AUTH"); ok && s.email != "" && s.password != "" {
		auth := smtp.PlainAuth("", s.email, s.password, s.host)
		if err = c.Auth(auth); err != nil {
			c.Close()
			conn.Close()
			return nil, nil, err
		}
	}

	return c, conn, nil
}
