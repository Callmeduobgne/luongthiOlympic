package chaincode

// TeaBatch represents the tea batch data model from teaTraceCC
// TeaBatch represents the tea batch data model from teaTraceCC
type TeaBatch struct {
	BatchID          string `json:"batchId"`
	FarmLocation     string `json:"farmLocation"`
	HarvestDate      string `json:"harvestDate"`
	ProcessingInfo   string `json:"processingInfo"`
	QualityCert      string `json:"qualityCert"`
	Status           string `json:"status"`
	VerificationHash string `json:"hashValue"` // Mapped to hashValue from chaincode
	Owner            string `json:"owner"`
	Timestamp        int64  `json:"timestamp"` // Chaincode uses number (milliseconds)
}

// TeaPackage represents the tea package data model from teaTraceCC
type TeaPackage struct {
	PackageID      string  `json:"packageId"`
	BatchID        string  `json:"batchId"`
	BlockHash      string  `json:"blockHash"`
	TxID           string  `json:"txId"`
	Weight         float64 `json:"weight"`
	ProductionDate string  `json:"productionDate"`
	ExpiryDate     string  `json:"expiryDate,omitempty"` // Changed to string (not pointer) for easier usage
	QRCode         string  `json:"qrCode,omitempty"`     // Changed to string
	Status         string  `json:"status"`
	Owner          string  `json:"owner"`
	Timestamp      int64   `json:"timestamp"`
}
