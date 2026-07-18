package inbound

import (
	"context"
	"io"
)

// ImportEmailsUseCase is the inbound driving port for importing emails from an mbox source.
type ImportEmailsUseCase interface {
	ImportFromMbox(ctx context.Context, r io.Reader) (int, error)
}
