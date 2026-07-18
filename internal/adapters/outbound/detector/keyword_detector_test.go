package detector_test

import (
	"context"
	"testing"

	"takeout_services/internal/adapters/outbound/detector"
	"takeout_services/internal/domain/model"
)

func TestKeywordDetector(t *testing.T) {
	emails := []*model.Email{
		{
			From:     "GitHub <noreply@github.com>",
			Subject:  "Welcome to GitHub!",
			BodyText: "Thanks for registering with us.",
		},
		{
			From:     "GitHub <noreply@github.com>",
			Subject:  "Payment Receipt for invoice 123",
			BodyText: "Your subscription has been renewed.",
		},
		{
			From:     "Netflix <info@netflix.com>",
			Subject:  "Password Reset Request",
			BodyText: "Reset your Netflix password here.",
		},
		{
			From:     "Spam Service <spam@spambot.xyz>",
			Subject:  "Buy cheap stuff",
			BodyText: "Ad content here.",
		},
	}

	det := detector.NewKeywordDetector()
	results, err := det.Detect(context.Background(), emails)
	if err != nil {
		t.Fatalf("Detection failed: %v", err)
	}

	// We expect github.com, netflix.com, and spambot.xyz to be detected (since unclassified services are included).
	if len(results) != 3 {
		t.Fatalf("Expected 3 detected services, got %d", len(results))
	}

	var github, netflix, spambot *model.DetectedService
	for _, res := range results {
		if res.Domain == "github.com" {
			github = res
		} else if res.Domain == "netflix.com" {
			netflix = res
		} else if res.Domain == "spambot.xyz" {
			spambot = res
		}
	}

	if github == nil {
		t.Fatal("GitHub not detected")
	}
	if !github.HasWelcome || !github.HasReceipt || github.HasReset {
		t.Errorf("GitHub indicators incorrect: %+v", github)
	}
	if github.Name != "GitHub" || github.DeleteURL != "https://github.com/settings/security" {
		t.Errorf("GitHub metadata mapping incorrect: %+v", github)
	}
	if github.Confidence < 9 { // 5 (welcome) + 4 (receipt) = 9
		t.Errorf("GitHub confidence score too low: %d", github.Confidence)
	}

	if netflix == nil {
		t.Fatal("Netflix not detected")
	}
	if netflix.HasWelcome || netflix.HasReceipt || !netflix.HasReset {
		t.Errorf("Netflix indicators incorrect: %+v", netflix)
	}
	if netflix.Name != "Netflix" || netflix.DeleteURL != "https://www.netflix.com/CancelPlan" {
		t.Errorf("Netflix metadata mapping incorrect: %+v", netflix)
	}

	if spambot == nil {
		t.Fatal("Spambot not detected")
	}
	if spambot.HasWelcome || spambot.HasReset || spambot.HasReceipt {
		t.Errorf("Spambot should not have welcome/reset/receipt indicators: %+v", spambot)
	}
	if spambot.Confidence != 1 {
		t.Errorf("Spambot confidence should be 1, got %d", spambot.Confidence)
	}
}
