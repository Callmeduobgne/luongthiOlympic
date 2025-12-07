package chaincode

// InstalledChaincode represents an installed chaincode
type InstalledChaincode struct {
	PackageID string `json:"packageId"`
	Label     string `json:"label"`
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
	Name                string   `json:"name"`
	Version             string   `json:"version"`
	Sequence            int64    `json:"sequence"`
	EndorsementPlugin   string   `json:"endorsementPlugin"`
	ValidationPlugin    string   `json:"validationPlugin"`
	InitRequired        bool     `json:"initRequired"`
	Collections         []string `json:"collections,omitempty"`
	ApprovedOrganizations []string `json:"approvedOrganizations"`
}

// ApproveChaincodeRequest represents a chaincode approval request
type ApproveChaincodeRequest struct {
	ChannelName         string   `json:"channelName"` // Channel name for approval
	Name                string   `json:"name"`
	Version             string   `json:"version"`
	Sequence            int64    `json:"sequence"`
	PackageID           string   `json:"packageId,omitempty"`
	InitRequired        bool     `json:"initRequired"`
	EndorsementPlugin   string   `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string   `json:"validationPlugin,omitempty"`
	Collections         []string `json:"collections,omitempty"`
}

// CommitChaincodeRequest represents a chaincode commit request
type CommitChaincodeRequest struct {
	ChannelName         string   `json:"channelName"` // Channel name for commit
	Name                string   `json:"name"`
	Version             string   `json:"version"`
	Sequence            int64    `json:"sequence"`
	InitRequired        bool     `json:"initRequired"`
	EndorsementPlugin   string   `json:"endorsementPlugin,omitempty"`
	ValidationPlugin    string   `json:"validationPlugin,omitempty"`
	Collections         []string `json:"collections,omitempty"`
}

