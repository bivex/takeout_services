package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"takeout_services/internal/domain/model"
)

// PrintEmailDetails reads a JSONL file and prints the subject and body of the target email.
func PrintEmailDetails(jsonlPath string, target string) error {
	f, err := os.Open(jsonlPath)
	if err != nil {
		return fmt.Errorf("could not open database file %s: %v. Please make sure to import emails first using: ./takeout-parser --input <mbox>", jsonlPath, err)
	}
	defer f.Close()

	// Parse if target is an index
	targetIdx, err := strconv.Atoi(target)
	isIndex := (err == nil)

	scanner := bufio.NewScanner(f)
	// Set scanner buffer size to handle potentially very large emails
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024) // up to 10MB per line

	currentIndex := 0
	found := false

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Quick check before parsing JSON to speed up lookups:
		// If we're searching by index, we don't need to parse other lines at all!
		if isIndex && currentIndex != targetIdx {
			currentIndex++
			continue
		}

		var email model.Email
		if err := json.Unmarshal(line, &email); err != nil {
			return fmt.Errorf("error parsing JSON on line %d: %v", currentIndex+1, err)
		}

		match := false
		if isIndex {
			match = (currentIndex == targetIdx)
		} else {
			match = (email.ID == target || email.MessageID == target || (len(target) >= 6 && strings.HasPrefix(email.ID, target)))
		}

		if match {
			// Print email details
			fmt.Println("================================================================================")
			fmt.Printf("Index:    %d\n", currentIndex)
			fmt.Printf("ID:       %s\n", email.ID)
			fmt.Printf("MsgID:    %s\n", email.MessageID)
			fmt.Printf("From:     %s\n", email.From)
			fmt.Printf("Date:     %s\n", email.Date.Format("2006-01-02 15:04:05"))
			fmt.Printf("Subject:  %s\n", email.Subject)
			fmt.Println("================================================================================")
			if email.BodyText != "" {
				fmt.Println(email.BodyText)
			} else {
				fmt.Println("[No Text Body, HTML content only:]")
				fmt.Println(email.BodyHTML)
			}
			fmt.Println("================================================================================")
			found = true
			break
		}

		currentIndex++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading database: %v", err)
	}

	if !found {
		if isIndex {
			return fmt.Errorf("email index %d out of bounds (database has %d emails)", targetIdx, currentIndex)
		}
		return fmt.Errorf("email with ID/MsgID '%s' not found in database", target)
	}

	return nil
}
