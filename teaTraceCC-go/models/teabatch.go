package models

// TeaBatchStatus represents the status of a tea batch
type TeaBatchStatus string

const (
	BatchCreated  TeaBatchStatus = "CREATED"
	BatchVerified TeaBatchStatus = "VERIFIED"
	BatchExpired  TeaBatchStatus = "EXPIRED"
)

// IsValidBatchStatus checks if a status is valid
func IsValidBatchStatus(status string) bool {
	switch TeaBatchStatus(status) {
	case BatchCreated, BatchVerified, BatchExpired:
		return true
	}
	return false
}

// TeaBatch represents a tea batch in the ledger
type TeaBatch struct {
	DocType        string         `json:"docType"`
	BatchID        string         `json:"batchId"`
	FarmLocation   string         `json:"farmLocation"`
	HarvestDate    string         `json:"harvestDate"`
	ProcessingInfo string         `json:"processingInfo"`
	QualityCert    string         `json:"qualityCert"`
	HashValue      string         `json:"hashValue"`
	Owner          string         `json:"owner"`
	Timestamp      string         `json:"timestamp"`
	Status         TeaBatchStatus `json:"status"`
}

// CreateTeaBatchInput represents input for creating a tea batch
type CreateTeaBatchInput struct {
	BatchID        string `json:"batchId"`
	FarmLocation   string `json:"farmLocation"`
	HarvestDate    string `json:"harvestDate"`
	ProcessingInfo string `json:"processingInfo"`
	QualityCert    string `json:"qualityCert"`
}
