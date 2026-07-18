package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"takeout_services/internal/adapters/inbound/server"
	"takeout_services/internal/adapters/outbound/detector"
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
	cpuprofile := flag.String("cpuprofile", "", "Write CPU profile to specified file")
	memprofile := flag.String("memprofile", "", "Write memory profile to specified file")

	// Footprint Detector flags
	detect := flag.Bool("detect", false, "Run digital footprint detection on the email archive")
	reportJSON := flag.String("report-json", "footprint.json", "Path to write the detected services JSON report")
	reportHTML := flag.String("report-html", "report.html", "Path to write the visual HTML report dashboard")

	// Server flags
	serve := flag.Bool("serve", false, "Start a local Go web server to host the interactive report and sync checkbox states")
	port := flag.Int("port", 8000, "Port to run the Go web server on")

	// Email reader flag
	readEmail := flag.String("read-email", "", "Read the subject and body of an email by its index (e.g., '5') or ID/MsgID from the emails JSONL database")

	flag.Parse()

	if *readEmail != "" {
		if err := services.PrintEmailDetails(*outputPath, *readEmail); err != nil {
			log.Fatalf("Error: %v", err)
		}
		return
	}

	if *inputPath == "" && !*serve {
		fmt.Println("Error: Either --input, --serve, or --read-email flag must be specified.")
		flag.Usage()
		os.Exit(1)
	}

	// Start CPU profile if requested
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatalf("could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	ctx := context.Background()
	startTime := time.Now()

	var mboxFile *os.File
	var err error
	if *inputPath != "" {
		// Open input mbox file
		mboxFile, err = os.Open(*inputPath)
		if err != nil {
			log.Fatalf("Error opening input mbox file: %v", err)
		}
		defer mboxFile.Close()
	}

	if *detect && mboxFile != nil {
		// Run Footprint Detection Usecase
		if *verbose {
			fmt.Printf("Analyzing digital footprint from: %s\n", *inputPath)
		}

		parser := mbox.NewParser()
		det := detector.NewKeywordDetector()
		analyzer := services.NewFootprintAnalyzer(parser, det)

		count, err := analyzer.AnalyzeFootprint(ctx, mboxFile, *reportJSON, *reportHTML)
		if err != nil {
			log.Fatalf("Error analyzing footprint: %v", err)
		}

		if *verbose {
			fmt.Printf("Footprint analysis complete. Detected %d unique services.\n", count)
			if *reportJSON != "" {
				fmt.Printf("JSON report written to: %s\n", *reportJSON)
			}
			if *reportHTML != "" {
				fmt.Printf("HTML dashboard report written to: %s\n", *reportHTML)
			}
		}
	} else if mboxFile != nil {
		// Run Standard Email Importing Usecase
		parser := mbox.NewParser()
		jsonRepo, err := repository.NewJSONLinesRepository(*outputPath)
		if err != nil {
			log.Fatalf("Error creating output repository: %v", err)
		}
		defer jsonRepo.Close()

		progressRepo := &progressRepository{
			next:    jsonRepo,
			verbose: *verbose,
		}

		if *verbose {
			fmt.Printf("Analyzing and parsing: %s\n", *inputPath)
			fmt.Printf("Writing output to: %s\n", *outputPath)
		}

		importer := services.NewEmailImporter(parser, progressRepo)

		count, err := importer.ImportFromMbox(ctx, mboxFile)
		if err != nil {
			log.Fatalf("\nError importing emails: %v", err)
		}

		if *verbose {
			fmt.Printf("\rProcessed %d emails... Done!\n", count)
			fmt.Printf("Successfully imported %d emails in %v\n", count, time.Since(startTime))
		}
	}

	// Write memory profile if requested
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatalf("could not create memory profile: %v", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("could not write memory profile: %v", err)
		}
	}

	if *serve {
		// Start Go HTTP Web Server
		if err := server.StartServer(*port, *reportHTML, "deleted_services.json"); err != nil {
			log.Fatalf("Error starting web server: %v", err)
		}
	}
}
