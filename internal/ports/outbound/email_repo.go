package outbound

import (
	"context"
	"takeout_services/internal/domain/model"
)

// EmailRepository is the outbound driven port for persisting emails.
type EmailRepository interface {
	Save(ctx context.Context, email *model.Email) error
}
