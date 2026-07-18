package model

import (
	"time"
)

// Attachment represents a file attachment in the email.
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Data        []byte `json:"-"`
}

// Email represents the core Domain Entity for an email message.
type Email struct {
	ID          string            `json:"id"`
	MessageID   string            `json:"message_id"`
	From        string            `json:"from"`
	To          []string          `json:"to"`
	Cc          []string          `json:"cc"`
	Bcc         []string          `json:"bcc"`
	Subject     string            `json:"subject"`
	Date        time.Time         `json:"date"`
	RawDate     string            `json:"raw_date"`
	BodyText    string            `json:"body_text"`
	BodyHTML    string            `json:"body_html"`
	Headers     map[string]string `json:"headers"`
	Attachments []Attachment      `json:"attachments"`
}

// NewEmail creates a new email instance.
func NewEmail(id, messageID, from string, to, cc, bcc []string, subject string, date time.Time, rawDate, bodyText, bodyHTML string, headers map[string]string, attachments []Attachment) *Email {
	return &Email{
		ID:          id,
		MessageID:   messageID,
		From:        from,
		To:          to,
		Cc:          cc,
		Bcc:         bcc,
		Subject:     subject,
		Date:        date,
		RawDate:     rawDate,
		BodyText:    bodyText,
		BodyHTML:    bodyHTML,
		Headers:     headers,
		Attachments: attachments,
	}
}
