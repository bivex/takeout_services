package services_test

import (
	"context"
	"strings"
	"testing"

	"takeout_services/internal/adapters/outbound/mbox"
	"takeout_services/internal/adapters/outbound/repository"
	"takeout_services/internal/application/services"
)

func TestEmailImporter(t *testing.T) {
	mboxContent := `From sender@example.com Sat Jul 18 08:00:00 2026
From: Sender <sender@example.com>
To: Recipient <recipient@example.com>
Subject: Test Email 1
Date: Sat, 18 Jul 2026 08:00:00 +0000
Message-ID: <test-1@example.com>
Content-Type: text/plain

Body text 1

From sender2@example.com Sat Jul 18 09:00:00 2026
From: Sender 2 <sender2@example.com>
To: Recipient 2 <recipient2@example.com>
Subject: Test Email 2
Date: Sat, 18 Jul 2026 09:00:00 +0000
Message-ID: <test-2@example.com>
Content-Type: text/plain

Body text 2
`

	parser := mbox.NewParser()
	repo := repository.NewInMemoryRepository()
	importer := services.NewEmailImporter(parser, repo)

	count, err := importer.ImportFromMbox(context.Background(), strings.NewReader(mboxContent))
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected to import 2 emails, got %d", count)
	}

	emails := repo.Emails()
	if len(emails) != 2 {
		t.Fatalf("Expected 2 emails in repo, got %d", len(emails))
	}

	if emails[0].Subject != "Test Email 1" || emails[0].BodyText != "Body text 1\n" {
		t.Errorf("Email 1 details incorrect: %+v", emails[0])
	}
	if emails[1].Subject != "Test Email 2" || emails[1].BodyText != "Body text 2\n" {
		t.Errorf("Email 2 details incorrect: %+v", emails[1])
	}
}
