package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"takeout_services/internal/adapters/outbound/mbox"
	"takeout_services/internal/adapters/outbound/repository"
	"takeout_services/internal/application/services"
	"takeout_services/internal/domain/model"
	"takeout_services/internal/ports/outbound"
)

// progressRepository wraps an outbound.EmailRepository to output progress to standard output.
type progressRepository struct {
	next    outbound.EmailRepository
	count   int
	verbose bool
}

func (p *progressRepository) Save(ctx context.Context, email *model.Email) error {
	err := p.next.Save(ctx, email)
	if err != nil {
		return err
	}
	p.count++
	if p.verbose && p.count%500 == 0 {
		fmt.Printf("\rProcessed %d emails...", p.count)
	}
	return nil
}

func main() {
	inputPath := flag.String("input", "", "Path to the input .mbox file")
	outputPath := flag.String("output", "emails.jsonl", "Path to write the output .jsonl file")
	verbose := flag.Bool("verbose", true, "Print parsing progress to stdout")
	flag.Parse()

	if *inputPath == "" {
		fmt.Println("Error: --input flag is required.")
		flag.Usage()
		os.Exit(1)
	}

	// 1. Open input mbox file
	mboxFile, err := os.Open(*inputPath)
	if err != nil {
		log.Fatalf("Error opening input mbox file: %v", err)
	}
	defer mboxFile.Close()

	// 2. Initialize outbound (driven) adapters
	parser := mbox.NewParser()
	jsonRepo, err := repository.NewJSONLinesRepository(*outputPath)
	if err != nil {
		log.Fatalf("Error creating output repository: %v", err)
	}
	defer jsonRepo.Close()

	// Wrap the repository to print progress on CLI
	progressRepo := &progressRepository{
		next:    jsonRepo,
		verbose: *verbose,
	}

	if *verbose {
		fmt.Printf("Analyzing and parsing: %s\n", *inputPath)
		fmt.Printf("Writing output to: %s\n", *outputPath)
	}

	// 3. Initialize application service with the driven ports
	importer := services.NewEmailImporter(parser, progressRepo)

	// 4. Run the use case
	ctx := context.Background()
	startTime := time.Now()

	count, err := importer.ImportFromMbox(ctx, mboxFile)
	if err != nil {
		log.Fatalf("\nError importing emails: %v", err)
	}

	duration := time.Since(startTime)
	if *verbose {
		// Clear/terminate progress log line and print summary
		fmt.Printf("\rProcessed %d emails... Done!\n", count)
		fmt.Printf("Successfully imported %d emails in %v\n", count, duration)
	}
}
