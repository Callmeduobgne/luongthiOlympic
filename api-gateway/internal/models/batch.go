package models

// TeaBatchStatus represents the status of a tea batch
type TeaBatchStatus string

const (
	StatusCreated  TeaBatchStatus = "CREATED"
	StatusVerified TeaBatchStatus = "VERIFIED"
	StatusExpired  TeaBatchStatus = "EXPIRED"
)

// TeaBatch represents a tea batch on the blockchain
type TeaBatch struct {
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

// CreateBatchRequest represents a request to create a new tea batch
type CreateBatchRequest struct {
	BatchID        string `json:"batchId" validate:"required,min=3,max=255"`
	FarmLocation   string `json:"farmLocation" validate:"required,min=3,max=255"`
	HarvestDate    string `json:"harvestDate" validate:"required,datetime=2006-01-02"`
	ProcessingInfo string `json:"processingInfo" validate:"required,min=10,max=1000"`
	QualityCert    string `json:"qualityCert" validate:"required,min=3,max=255"`
}

// VerifyBatchRequest represents a request to verify a batch
type VerifyBatchRequest struct {
	HashInput string `json:"hashInput" validate:"required"`
}

// VerifyBatchResponse represents the response from verifying a batch
type VerifyBatchResponse struct {
	IsValid bool     `json:"isValid"`
	Batch   TeaBatch `json:"batch"`
}

// UpdateBatchStatusRequest represents a request to update batch status
type UpdateBatchStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=CREATED VERIFIED EXPIRED"`
}

// IsValid checks if the status is valid
func (s TeaBatchStatus) IsValid() bool {
	switch s {
	case StatusCreated, StatusVerified, StatusExpired:
		return true
	}
	return false
}

// String returns the string representation of the status
func (s TeaBatchStatus) String() string {
	return string(s)
}

