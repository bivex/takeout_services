package mbox_test

import (
	"strings"
	"testing"

	"takeout_services/internal/adapters/outbound/mbox"
	"takeout_services/internal/domain/model"
)

func TestMboxParser(t *testing.T) {
	mboxContent := `From sender@example.com Sat Jul 18 08:00:00 2026
From: Sender <sender@example.com>
To: Recipient <recipient@example.com>
Subject: Test Email 1
Date: Sat, 18 Jul 2026 08:00:00 +0000
Message-ID: <test-1@example.com>
Content-Type: text/plain

This is the first test email body.

From sender2@example.com Sat Jul 18 09:00:00 2026
From: Sender 2 <sender2@example.com>
To: Recipient 2 <recipient2@example.com>
Subject: Test Email 2
Date: Sat, 18 Jul 2026 09:00:00 +0000
Message-ID: <test-2@example.com>
Content-Type: multipart/alternative; boundary="boundary-string"

--boundary-string
Content-Type: text/plain
Content-Transfer-Encoding: quoted-printable

Hello from the text part of email 2!

--boundary-string
Content-Type: text/html
Content-Transfer-Encoding: base64

SGVsbG8gZnJvbSB0aGUgSFRNTCBwYXJ0IG9mIGVtYWlsIDIhCg==
--boundary-string--
`

	parser := mbox.NewParser()
	var emails []*model.Email

	err := parser.Parse(strings.NewReader(mboxContent), func(email *model.Email) error {
		emails = append(emails, email)
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to parse mbox: %v", err)
	}

	if len(emails) != 2 {
		t.Fatalf("Expected 2 emails, got %d", len(emails))
	}

	// Verify email 1
	e1 := emails[0]
	if e1.Subject != "Test Email 1" {
		t.Errorf("Expected subject 'Test Email 1', got %q", e1.Subject)
	}
	if e1.BodyText != "This is the first test email body.\n" {
		t.Errorf("Expected body 'This is the first test email body.\n', got %q", e1.BodyText)
	}
	if e1.MessageID != "test-1@example.com" {
		t.Errorf("Expected message ID 'test-1@example.com', got %q", e1.MessageID)
	}

	// Verify email 2
	e2 := emails[1]
	if e2.Subject != "Test Email 2" {
		t.Errorf("Expected subject 'Test Email 2', got %q", e2.Subject)
	}
	if !strings.Contains(e2.BodyText, "Hello from the text part") {
		t.Errorf("Expected body text to contain 'Hello from the text part', got %q", e2.BodyText)
	}
	if !strings.Contains(e2.BodyHTML, "Hello from the HTML part") {
		t.Errorf("Expected HTML body to contain 'Hello from the HTML part', got %q", e2.BodyHTML)
	}
}
