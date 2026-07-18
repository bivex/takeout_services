package services

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"takeout_services/internal/adapters/outbound/report"
	"takeout_services/internal/domain/model"
	"takeout_services/internal/ports/inbound"
	"takeout_services/internal/ports/outbound"
)

// FootprintAnalyzer implements inbound.DetectFootprintUseCase.
type FootprintAnalyzer struct {
	parser   outbound.MboxParser
	detector outbound.ServiceDetector
}

// NewFootprintAnalyzer constructs a new FootprintAnalyzer.
func NewFootprintAnalyzer(parser outbound.MboxParser, detector outbound.ServiceDetector) inbound.DetectFootprintUseCase {
	return &FootprintAnalyzer{
		parser:   parser,
		detector: detector,
	}
}

// AnalyzeFootprint parses emails, runs the service detector, and writes reports to the target paths.
func (a *FootprintAnalyzer) AnalyzeFootprint(ctx context.Context, mboxReader io.Reader, jsonReportPath, htmlReportPath string) (int, error) {
	var emails []*model.Email
	err := a.parser.Parse(mboxReader, func(email *model.Email) error {
		emails = append(emails, email)
		return nil
	})
	if err != nil {
		return 0, err
	}

	detected, err := a.detector.Detect(ctx, emails)
	if err != nil {
		return 0, err
	}

	// 1. Output JSON report if requested
	if jsonReportPath != "" {
		f, err := os.Create(jsonReportPath)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(detected); err != nil {
			return 0, err
		}
	}

	// 2. Output HTML report if requested
	if htmlReportPath != "" {
		if err := report.GenerateHTMLReport(detected, htmlReportPath); err != nil {
			return 0, err
		}
	}

	return len(detected), nil
}
