package gmgo

type Email struct {
	To          []string     `json:"to"`
	Cc          []string     `json:"cc,omitempty"`
	Bcc         []string     `json:"bcc,omitempty"`
	Subject     string       `json:"subject"`
	Body        string       `json:"body"`
	IsHTML      bool         `json:"is_html,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	From        string       `json:"from,omitempty"`
	ReplyTo     string       `json:"reply_to,omitempty"`
}

// Attachment representa un archivo adjunto
type Attachment struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
	MimeType string `json:"mime_type"`
}

func (email *Email) AddRecipient(to string) *Email {
	email.To = append(email.To, to)
	return email
}

func (email *Email) AttachFile(filename string, content []byte, mimeType ...string) *Email {
	attachment := Attachment{
		Filename: filename,
		Content:  content,
	}

	if len(mimeType) > 0 {
		attachment.MimeType = mimeType[0]
	}

	email.Attachments = append(email.Attachments, attachment)

	return email
}
