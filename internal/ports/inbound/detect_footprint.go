package inbound

import (
	"context"
	"io"
)

// DetectFootprintUseCase is the inbound driving port for analyzing the digital footprint from an mbox stream.
type DetectFootprintUseCase interface {
	AnalyzeFootprint(ctx context.Context, mboxReader io.Reader, jsonReportPath, htmlReportPath string) (int, error)
}
