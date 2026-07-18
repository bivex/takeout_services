package outbound

import (
	"io"
	"takeout_services/internal/domain/model"
)

// MboxParser is the outbound driven port for parsing emails from an mbox stream.
type MboxParser interface {
	Parse(r io.Reader, callback func(*model.Email) error) error
}
