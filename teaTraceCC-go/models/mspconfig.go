package models

// MSPConfig holds the MSP IDs for different roles
type MSPConfig struct {
	Farmer   string
	Verifier string
	Admin    string
}

// DefaultMSPConfig returns the default MSP configuration
func DefaultMSPConfig() MSPConfig {
	return MSPConfig{
		Farmer:   "Org1MSP",
		Verifier: "Org1MSP",
		Admin:    "Org1MSP",
	}
}
