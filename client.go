package gmgo

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GmailClient struct {
	service *gmail.Service
	ctx     context.Context
	config  *Config
}

func NewClient(cfg *Config) (*GmailClient, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{gmail.GmailSendScope}
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	oauthConfig, token, err := loadConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}

	ctx := context.Background()
	httpClient := oauthConfig.Client(context.Background(), token)
	service, err := gmail.NewService(ctx, option.WithHTTPClient(httpClient))

	if err != nil {
		return nil, fmt.Errorf("error creating gmail service: %w", err)
	}

	return &GmailClient{
		service: service,
		ctx:     ctx,
		config:  cfg,
	}, nil
}

func DefaultConfig() *Config {
	return &Config{
		CredentialsFile: "gmgo_credentials.json",
		TokenFile:       "gmgo_token.json",
		Scopes:          []string{gmail.GmailSendScope},
	}
}

// SendResult contains the result of the send operation
type SendResult struct {
	MessageID string    `json:"message_id"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	SentAt    time.Time `json:"sent_at"`
	ThreadID  string    `json:"thread_id,omitempty"`
}

// SendEmail sends an individual email
func (c *GmailClient) SendEmail(email *Email) (*SendResult, error) {
	result := &SendResult{
		SentAt: time.Now(),
	}

	if err := c.validateEmail(email); err != nil {
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	message, err := c.createMessage(email)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("error creando mensaje: %v", err)
		return result, err
	}

	sent, err := c.service.Users.Messages.Send("me", message).Do()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("error enviando email: %v", err)
		return result, err
	}

	result.MessageID = sent.Id
	result.ThreadID = sent.ThreadId
	result.Success = true

	return result, nil
}

func (c *GmailClient) validateEmail(email *Email) error {
	if len(email.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	if email.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if email.Body == "" {
		return fmt.Errorf("email body is required")
	}

	// Validate email addresses
	allEmails := append(email.To, email.Cc...)
	allEmails = append(allEmails, email.Bcc...)

	for _, emailAddr := range allEmails {
		if !strings.Contains(emailAddr, "@") {
			return fmt.Errorf("invalid email: %s", emailAddr)
		}
	}

	return nil
}

func (c *GmailClient) createMessage(email *Email) (*gmail.Message, error) {
	var buf bytes.Buffer

	if len(email.Attachments) > 0 {
		return c.createMessageWithAttachments(email)
	}

	// Basic headers
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))

	if len(email.Cc) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.Cc, ", ")))
	}

	if len(email.Bcc) > 0 {
		buf.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(email.Bcc, ", ")))
	}

	if email.From != "" {
		buf.WriteString(fmt.Sprintf("From: %s\r\n", email.From))
	}

	if email.ReplyTo != "" {
		buf.WriteString(fmt.Sprintf("Reply-To: %s\r\n", email.ReplyTo))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))

	if email.IsHTML {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	buf.WriteString("\r\n")
	buf.WriteString(email.Body)

	// Encode in base64
	raw := base64.URLEncoding.EncodeToString(buf.Bytes())

	return &gmail.Message{Raw: raw}, nil
}

func (c *GmailClient) createMessageWithAttachments(email *Email) (*gmail.Message, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	boundary := writer.Boundary()

	header := make(textproto.MIMEHeader)
	header.Set("To", strings.Join(email.To, ", "))

	if len(email.Cc) > 0 {
		header.Set("Cc", strings.Join(email.Cc, ", "))
	}

	if len(email.Bcc) > 0 {
		header.Set("Bcc", strings.Join(email.Bcc, ", "))
	}

	if email.From != "" {
		header.Set("From", email.From)
	}

	header.Set("Subject", email.Subject)
	header.Set("Date", time.Now().Format(time.RFC1123Z))
	header.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", boundary))

	for k, v := range header {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ", ")))
	}
	buf.WriteString("\r\n")

	// Message body
	bodyHeader := make(textproto.MIMEHeader)
	if email.IsHTML {
		bodyHeader.Set("Content-Type", "text/html; charset=UTF-8")
	} else {
		bodyHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	}

	bodyWriter, err := writer.CreatePart(bodyHeader)
	if err != nil {
		return nil, fmt.Errorf("error creando parte del cuerpo: %w", err)
	}

	if _, err := bodyWriter.Write([]byte(email.Body)); err != nil {
		return nil, fmt.Errorf("error escribiendo cuerpo: %w", err)
	}

	// Attachments
	for _, attachment := range email.Attachments {
		if err := c.addAttachment(writer, &attachment); err != nil {
			return nil, fmt.Errorf("error agregando archivo adjunto %s: %w", attachment.Filename, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error cerrando writer: %w", err)
	}

	// Encode in base64
	raw := base64.URLEncoding.EncodeToString(buf.Bytes())

	return &gmail.Message{Raw: raw}, nil
}

func (c *GmailClient) addAttachment(writer *multipart.Writer, attachment *Attachment) error {
	// Determine MIME type if not specified
	mimeType := attachment.MimeType
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(attachment.Filename))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	header := make(textproto.MIMEHeader)
	header.Set("Content-Type", mimeType)
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", attachment.Filename))
	header.Set("Content-Transfer-Encoding", "base64")

	partWriter, err := writer.CreatePart(header)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(attachment.Content)
	_, err = partWriter.Write([]byte(encoded))

	return err
}
