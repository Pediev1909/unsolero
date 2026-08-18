package email

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	"rigmark/internal/modules/identity/ports"
)

type SMTPConfig struct {
	Address       string
	Username      string
	Password      string
	SenderName    string
	SenderAddress string
	PublicSiteURL string
	RequireTLS    bool
	Timeout       time.Duration
}

type OutboundMessage struct {
	Recipient string
	Subject   string
	TextBody  string
}

type Sender interface {
	Send(context.Context, OutboundMessage) error
}

type SMTPDelivery struct {
	sender  Sender
	siteURL *url.URL
}

func NewSMTPDelivery(config SMTPConfig) (*SMTPDelivery, error) {
	sender, err := newSMTPSender(config)
	if err != nil {
		return nil, err
	}
	return NewSMTPDeliveryWithSender(sender, config.PublicSiteURL)
}

func NewSMTPDeliveryWithSender(sender Sender, publicSiteURL string) (*SMTPDelivery, error) {
	siteURL, err := url.Parse(publicSiteURL)
	if sender == nil || err != nil || siteURL.Scheme == "" || siteURL.Host == "" || siteURL.RawQuery != "" || siteURL.Fragment != "" {
		return nil, errors.New("email delivery requires a sender and safe public site URL")
	}
	return &SMTPDelivery{sender: sender, siteURL: siteURL}, nil
}

func (delivery *SMTPDelivery) SendVerification(ctx context.Context, message ports.VerificationMessage) (ports.DeliveryReceipt, error) {
	link, err := delivery.tokenURL("/verify-email", message.Token)
	if err != nil || message.ExpiresAt.IsZero() {
		return ports.DeliveryReceipt{}, errors.New("invalid verification message")
	}
	return delivery.send(ctx, message.Recipient, "Verify your UNSOLERO email", "Verify your email to finish securing your UNSOLERO account.\n\n"+link+"\n\nThis one-time link expires at "+message.ExpiresAt.UTC().Format(time.RFC3339)+". If you did not request this, ignore this message.")
}

func (delivery *SMTPDelivery) SendPasswordReset(ctx context.Context, message ports.PasswordResetMessage) (ports.DeliveryReceipt, error) {
	link, err := delivery.tokenURL("/reset-password", message.Token)
	if err != nil || message.ExpiresAt.IsZero() {
		return ports.DeliveryReceipt{}, errors.New("invalid password-reset message")
	}
	return delivery.send(ctx, message.Recipient, "Reset your UNSOLERO password", "A password reset was requested for your UNSOLERO account.\n\n"+link+"\n\nThis one-time link expires at "+message.ExpiresAt.UTC().Format(time.RFC3339)+". If you did not request this, secure your account and contact support.")
}

func (delivery *SMTPDelivery) SendSecurityNotification(ctx context.Context, message ports.SecurityNotification) (ports.DeliveryReceipt, error) {
	description, allowed := map[string]string{
		"password_changed":         "Your UNSOLERO password was changed.",
		"mfa_enabled":              "Multi-factor authentication was enabled for your UNSOLERO account.",
		"mfa_recovery_regenerated": "New MFA recovery codes were generated for your UNSOLERO account.",
	}[message.EventType]
	if !allowed || message.OccurredAt.IsZero() {
		return ports.DeliveryReceipt{}, errors.New("invalid security notification")
	}
	body := description + "\n\nTime: " + message.OccurredAt.UTC().Format(time.RFC3339) + "\n\nIf this was not you, reset your password and contact support immediately."
	return delivery.send(ctx, message.Recipient, "UNSOLERO account security notice", body)
}

func (delivery *SMTPDelivery) send(ctx context.Context, recipient, subject, body string) (ports.DeliveryReceipt, error) {
	if _, err := parseMailbox(recipient); err != nil {
		return ports.DeliveryReceipt{}, errors.New("invalid email recipient")
	}
	if err := delivery.sender.Send(ctx, OutboundMessage{Recipient: recipient, Subject: subject, TextBody: body}); err != nil {
		return ports.DeliveryReceipt{Accepted: false}, fmt.Errorf("email transport rejected message: %w", err)
	}
	return ports.DeliveryReceipt{Accepted: true, Reference: "smtp_accepted"}, nil
}

func (delivery *SMTPDelivery) tokenURL(path, token string) (string, error) {
	if len(token) < 32 || len(token) > 512 || strings.ContainsAny(token, "\r\n?#&/") {
		return "", errors.New("invalid one-time token")
	}
	target := *delivery.siteURL
	target.Path = path
	target.RawQuery = ""
	target.Fragment = token
	return target.String(), nil
}

type smtpSender struct {
	config SMTPConfig
	from   *mail.Address
	host   string
}

func newSMTPSender(config SMTPConfig) (*smtpSender, error) {
	if config.Timeout < time.Second || config.Timeout > time.Minute {
		return nil, errors.New("SMTP timeout must be between one second and one minute")
	}
	host, _, err := net.SplitHostPort(config.Address)
	if err != nil || host == "" || strings.ContainsAny(config.Address, "\r\n") {
		return nil, errors.New("SMTP address must be host:port")
	}
	from, err := parseMailbox((&mail.Address{Name: config.SenderName, Address: config.SenderAddress}).String())
	if err != nil || (config.Username == "") != (config.Password == "") {
		return nil, errors.New("SMTP sender or credentials are invalid")
	}
	return &smtpSender{config: config, from: from, host: host}, nil
}

func (sender *smtpSender) Send(ctx context.Context, message OutboundMessage) error {
	recipient, err := parseMailbox(message.Recipient)
	if err != nil || strings.ContainsAny(message.Subject, "\r\n") {
		return errors.New("invalid SMTP message")
	}
	dialer := net.Dialer{Timeout: sender.config.Timeout}
	connection, err := dialer.DialContext(ctx, "tcp", sender.config.Address)
	if err != nil {
		return err
	}
	defer connection.Close()
	deadline := time.Now().Add(sender.config.Timeout)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	_ = connection.SetDeadline(deadline)
	client, err := smtp.NewClient(connection, sender.host)
	if err != nil {
		return err
	}
	defer client.Close()
	if supported, _ := client.Extension("STARTTLS"); supported {
		if err = client.StartTLS(&tls.Config{ServerName: sender.host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	} else if sender.config.RequireTLS {
		return errors.New("SMTP server does not support required STARTTLS")
	}
	if sender.config.Username != "" {
		if ok, _ := client.Extension("AUTH"); !ok {
			return errors.New("SMTP server does not support authentication")
		}
		if err = client.Auth(smtp.PlainAuth("", sender.config.Username, sender.config.Password, sender.host)); err != nil {
			return err
		}
	}
	if err = client.Mail(sender.from.Address); err != nil {
		return err
	}
	if err = client.Rcpt(recipient.Address); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	buffer := bufio.NewWriter(writer)
	_, err = fmt.Fprintf(buffer, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n", sender.from.String(), recipient.String(), message.Subject, normalizeBody(message.TextBody))
	if flushErr := buffer.Flush(); err == nil {
		err = flushErr
	}
	if closeErr := writer.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = client.Quit(); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func parseMailbox(value string) (*mail.Address, error) {
	if strings.ContainsAny(value, "\r\n") {
		return nil, errors.New("mailbox contains a line break")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address == "" || strings.ContainsAny(address.Address, "\r\n") {
		return nil, errors.New("invalid mailbox")
	}
	return address, nil
}

func normalizeBody(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.ReplaceAll(value, "\n", "\r\n")
}
