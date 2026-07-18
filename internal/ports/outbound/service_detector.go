package outbound

import (
	"context"
	"takeout_services/internal/domain/model"
)

// ServiceDetector is the outbound driven port for detecting services from email lists.
type ServiceDetector interface {
	Detect(ctx context.Context, emails []*model.Email) ([]*model.DetectedService, error)
}
