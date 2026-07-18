package detector

import (
	"context"
	"fmt"
	"strings"

	"takeout_services/internal/domain/model"
	"takeout_services/internal/ports/outbound"
)

// KeywordDetector implements outbound.ServiceDetector using email analysis.
type KeywordDetector struct{}

// NewKeywordDetector creates a new KeywordDetector.
func NewKeywordDetector() outbound.ServiceDetector {
	return &KeywordDetector{}
}

func (d *KeywordDetector) Detect(ctx context.Context, emails []*model.Email) ([]*model.DetectedService, error) {
	type serviceAccumulator struct {
		domain         string
		hasWelcome     bool
		hasReset       bool
		hasReceipt     bool
		sourcesCount   int
		sampleSubjects []string
	}

	accums := make(map[string]*serviceAccumulator)

	// Keywords for classification (case-insensitive checks)
	welcomeKeywords := []string{
		"welcome", "verify email", "confirm account", "registration", "activated", "started with",
		"создан", "добро пожаловать", "подтвердите", "активация", "регистрация", "верификац",
	}

	resetKeywords := []string{
		"password reset", "reset your password", "сброс пароля", "восстановление пароля", "парол",
	}

	receiptKeywords := []string{
		"receipt", "invoice", "payment", "subscription", "order", "purchase",
		"чек", "оплата", "квитанция", "подписка", "покупка",
	}

	containsAny := func(text string, keywords []string) bool {
		lowerText := strings.ToLower(text)
		for _, kw := range keywords {
			if strings.Contains(lowerText, kw) {
				return true
			}
		}
		return false
	}

	for _, email := range emails {
		senderDomain := extractDomain(email.From)
		if senderDomain == "" {
			continue
		}

		baseDomain := getBaseDomain(senderDomain)
		if baseDomain == "" {
			continue
		}

		// Skip common personal email providers from being detected as unique services
		if isCommonProvider(baseDomain) {
			continue
		}

		accum, exists := accums[baseDomain]
		if !exists {
			accum = &serviceAccumulator{
				domain: baseDomain,
			}
			accums[baseDomain] = accum
		}

		accum.sourcesCount++

		// Check keywords in Subject or Body
		contentToCheck := email.Subject + " " + email.BodyText
		if !accum.hasWelcome && containsAny(contentToCheck, welcomeKeywords) {
			accum.hasWelcome = true
		}
		if !accum.hasReset && containsAny(contentToCheck, resetKeywords) {
			accum.hasReset = true
		}
		if !accum.hasReceipt && containsAny(contentToCheck, receiptKeywords) {
			accum.hasReceipt = true
		}

		// Keep up to 3 sample subjects
		if len(accum.sampleSubjects) < 3 {
			// Avoid duplicates in subjects
			duplicate := false
			for _, sub := range accum.sampleSubjects {
				if sub == email.Subject {
					duplicate = true
					break
				}
			}
			if !duplicate && email.Subject != "" {
				accum.sampleSubjects = append(accum.sampleSubjects, email.Subject)
			}
		}
	}

	var results []*model.DetectedService
	for baseDomain, accum := range accums {
		name := capitalizeDomain(baseDomain)
		deleteURL := fmt.Sprintf("https://www.google.com/search?q=%s+delete+account", baseDomain)

		if metadata, ok := KnownServices[baseDomain]; ok {
			name = metadata.Name
			deleteURL = metadata.DeleteURL
		}

		// Scoring confidence:
		// +5 welcome/registration
		// +4 receipt/payment
		// +3 password reset
		// +1 per email (up to +3)
		confidence := 0
		if accum.hasWelcome {
			confidence += 5
		}
		if accum.hasReceipt {
			confidence += 4
		}
		if accum.hasReset {
			confidence += 3
		}

		sourceScore := accum.sourcesCount
		if sourceScore > 3 {
			sourceScore = 3
		}
		confidence += sourceScore

		if confidence > 10 {
			confidence = 10
		}

		// We filter out low confidence entries (e.g. confidence = 1 and only 1 email)
		// to avoid false positives from one-off spam domains.
		if confidence <= 1 && accum.sourcesCount <= 1 {
			continue
		}

		results = append(results, model.NewDetectedService(
			name,
			baseDomain,
			accum.hasWelcome,
			accum.hasReset,
			accum.hasReceipt,
			confidence,
			accum.sourcesCount,
			accum.sampleSubjects,
			deleteURL,
		))
	}

	return results, nil
}

func extractDomain(fromStr string) string {
	fromStr = strings.TrimSpace(fromStr)
	if fromStr == "" {
		return ""
	}
	idx := strings.Index(fromStr, "<")
	if idx != -1 {
		endIdx := strings.Index(fromStr[idx:], ">")
		if endIdx != -1 {
			fromStr = fromStr[idx+1 : idx+endIdx]
		}
	}
	atIdx := strings.Index(fromStr, "@")
	if atIdx != -1 {
		return strings.ToLower(strings.TrimSpace(fromStr[atIdx+1:]))
	}
	return strings.ToLower(strings.TrimSpace(fromStr))
}

func getBaseDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) <= 2 {
		return domain
	}

	last := parts[len(parts)-1]
	secondLast := parts[len(parts)-2]

	isDoubleExt := false
	if len(last) == 2 {
		switch secondLast {
		case "co", "com", "org", "net", "edu", "gov", "ac":
			isDoubleExt = true
		}
	}

	if isDoubleExt && len(parts) >= 3 {
		return strings.Join(parts[len(parts)-3:], ".")
	}
	return strings.Join(parts[len(parts)-2:], ".")
}

func isCommonProvider(domain string) bool {
	common := map[string]bool{
		"gmail.com":      true,
		"yahoo.com":      true,
		"hotmail.com":    true,
		"outlook.com":    true,
		"live.com":       true,
		"mail.ru":        true,
		"yandex.ru":      true,
		"yandex.ua":      true,
		"rambler.ru":     true,
		"icloud.com":     true,
		"protonmail.com": true,
		"proton.me":      true,
		"ukr.net":        true,
		"i.ua":           true,
	}
	return common[domain]
}

func capitalizeDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) > 0 {
		name := parts[0]
		if len(name) > 0 {
			return strings.ToUpper(name[:1]) + name[1:]
		}
	}
	return domain
}
