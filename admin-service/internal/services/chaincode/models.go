package chaincode

// InstalledChaincode represents an installed chaincode
type InstalledChaincode struct {
	PackageID string        `json:"packageId"`
	Label     string        `json:"label"`
	Chaincode ChaincodeInfo `json:"chaincode"`
}

// ChaincodeInfo represents chaincode information
type ChaincodeInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

// CommittedChaincode represents a committed chaincode on a channel
type CommittedChaincode struct {
	Name                 string   `json:"name"`
	Version              string   `json:"version"`
	Sequence             int64    `json:"sequence"`
	EndorsementPlugin    string   `json:"endorsementPlugin"`
	ValidationPlugin     string   `json:"validationPlugin"`
	InitRequired         bool     `json:"initRequired"`
	Collections          []string `json:"collections,omitempty"`
	ApprovedOrganizations []string `json:"approvedOrganizations"`
}

// InstallChaincodeRequest represents a chaincode installation request
type InstallChaincodeRequest struct {
	PackagePath string `json:"packagePath" validate:"required"`
	Label       string `json:"label,omitempty"`
}

// ApproveChaincodeRequest represents a chaincode approval request
type ApproveChaincodeRequest struct {
	ChannelName         string   `json:"channelName" validate:"required"`
	Name                string   `json:"name" validate:"required"`
	Version             string   `json:"version" validate:"required"`
	Sequence            int64    `json:"sequence" validate:"required,min=1"`
	PackageID           string   `json:"packageId,omitempty"`
	InitRequired        bool     `json:"initRequired"`
	EndorsementPlugin   string   `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string   `json:"validationPlugin,omitempty"`
	Collections         []string `json:"collections,omitempty"`
}

// CommitChaincodeRequest represents a chaincode commit request
type CommitChaincodeRequest struct {
	ChannelName         string   `json:"channelName" validate:"required"`
	Name                string   `json:"name" validate:"required"`
	Version             string   `json:"version" validate:"required"`
	Sequence            int64    `json:"sequence" validate:"required,min=1"`
	InitRequired        bool     `json:"initRequired"`
	EndorsementPlugin   string   `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string   `json:"validationPlugin,omitempty"`
	Collections         []string `json:"collections,omitempty"`
}

