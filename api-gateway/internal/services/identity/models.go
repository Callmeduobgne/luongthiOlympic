package identity

// UserInfo represents user information
type UserInfo struct {
	Username    string   `json:"username"`
	Type        string   `json:"type"`
	Affiliation string   `json:"affiliation"`
	Attributes  []string `json:"attributes"`
	Revoked     bool     `json:"revoked"`
}

