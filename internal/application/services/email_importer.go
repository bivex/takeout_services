package services

import (
	"context"
	"io"

	"takeout_services/internal/domain/model"
	"takeout_services/internal/ports/inbound"
	"takeout_services/internal/ports/outbound"
)

// EmailImporter orchestrates the email import usecase.
type EmailImporter struct {
	parser outbound.MboxParser
	repo   outbound.EmailRepository
}

// NewEmailImporter constructs a new EmailImporter.
func NewEmailImporter(parser outbound.MboxParser, repo outbound.EmailRepository) inbound.ImportEmailsUseCase {
	return &EmailImporter{
		parser: parser,
		repo:   repo,
	}
}

// ImportFromMbox reads emails using the parser port and saves them using the repository port.
func (s *EmailImporter) ImportFromMbox(ctx context.Context, r io.Reader) (int, error) {
	count := 0
	err := s.parser.Parse(r, func(email *model.Email) error {
		// Check for context cancellation before processing the next email
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := s.repo.Save(ctx, email); err != nil {
			return err
		}
		count++
		return nil
	})

	if err != nil {
		return count, err
	}

	return count, nil
}
