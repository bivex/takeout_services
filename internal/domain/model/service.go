package model

// DetectedService represents a digital service that the user has interacted with.
type DetectedService struct {
	Name           string   `json:"name"`
	Domain         string   `json:"domain"`
	HasWelcome     bool     `json:"has_welcome"`
	HasReset       bool     `json:"has_reset"`
	HasReceipt     bool     `json:"has_receipt"`
	Confidence     int      `json:"confidence"`
	SourcesCount   int      `json:"sources_count"`
	SampleSubjects []string `json:"sample_subjects"`
	DeleteURL      string   `json:"delete_url"`
}

// NewDetectedService constructs a new DetectedService.
func NewDetectedService(name, domain string, welcome, reset, receipt bool, confidence, count int, subjects []string, deleteURL string) *DetectedService {
	return &DetectedService{
		Name:           name,
		Domain:         domain,
		HasWelcome:     welcome,
		HasReset:       reset,
		HasReceipt:     receipt,
		Confidence:     confidence,
		SourcesCount:   count,
		SampleSubjects: subjects,
		DeleteURL:      deleteURL,
	}
}
