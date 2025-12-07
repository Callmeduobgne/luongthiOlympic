package models

// TeaPackageStatus represents the status of a tea package
type TeaPackageStatus string

const (
	PackageCreated  TeaPackageStatus = "CREATED"
	PackageVerified TeaPackageStatus = "VERIFIED"
	PackageSold     TeaPackageStatus = "SOLD"
	PackageExpired  TeaPackageStatus = "EXPIRED"
)

// IsValidPackageStatus checks if a status is valid
func IsValidPackageStatus(status string) bool {
	switch TeaPackageStatus(status) {
	case PackageCreated, PackageVerified, PackageSold, PackageExpired:
		return true
	}
	return false
}

// TeaPackage represents a tea package in the ledger
type TeaPackage struct {
	DocType        string           `json:"docType"`
	PackageID      string           `json:"packageId"`
	BatchID        string           `json:"batchId"`
	BlockHash      string           `json:"blockHash"`
	HashVersion    string           `json:"hashVersion,omitempty"`
	TxID           string           `json:"txId"`
	Weight         float64          `json:"weight"`
	ProductionDate string           `json:"productionDate"`
	ExpiryDate     string           `json:"expiryDate,omitempty"`
	QRCode         string           `json:"qrCode,omitempty"`
	Status         TeaPackageStatus `json:"status"`
	Owner          string           `json:"owner"`
	Timestamp      string           `json:"timestamp"`
}

// CreateTeaPackageInput represents input for creating a tea package
type CreateTeaPackageInput struct {
	PackageID      string  `json:"packageId"`
	BatchID        string  `json:"batchId"`
	Weight         float64 `json:"weight"`
	ProductionDate string  `json:"productionDate"`
	ExpiryDate     string  `json:"expiryDate,omitempty"`
	QRCode         string  `json:"qrCode,omitempty"`
}
