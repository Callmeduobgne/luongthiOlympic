package models

// CreateChannelRequest represents request to create a new channel
type CreateChannelRequest struct {
	Name         string   `json:"name" validate:"required,min=1"`
	Consortium   string   `json:"consortium" validate:"required"`
	Organizations []string `json:"organizations" validate:"required,min=1"`
	Profile      string   `json:"profile,omitempty"` // Configtx profile name
}

// CreateChannelResponse represents response after creating a channel
type CreateChannelResponse struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	GenesisBlock string `json:"genesisBlock,omitempty"` // Path to genesis block
	Message     string `json:"message"`
}

// UpdateChannelConfigRequest represents request to update channel config
type UpdateChannelConfigRequest struct {
	ConfigUpdate map[string]interface{} `json:"configUpdate" validate:"required"`
	Description  string                 `json:"description,omitempty"`
}

// UpdateChannelConfigResponse represents response after updating channel config
type UpdateChannelConfigResponse struct {
	ChannelName string `json:"channelName"`
	TxID        string `json:"txId,omitempty"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// JoinChannelRequest represents request to join a peer to channel
type JoinChannelRequest struct {
	PeerAddress string `json:"peerAddress" validate:"required"`
	BlockPath   string `json:"blockPath,omitempty"` // Path to genesis block
}

// JoinChannelResponse represents response after joining a peer to channel
type JoinChannelResponse struct {
	ChannelName string `json:"channelName"`
	PeerAddress string `json:"peerAddress"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// ChannelMember represents a member (organization) in a channel
type ChannelMember struct {
	MSPID         string   `json:"mspId"`
	Organization  string   `json:"organization"`
	Peers         []string `json:"peers"`
	AnchorPeers   []string `json:"anchorPeers,omitempty"`
	JoinedAt      string   `json:"joinedAt,omitempty"`
}

// ListChannelMembersResponse represents response for listing channel members
type ListChannelMembersResponse struct {
	ChannelName string          `json:"channelName"`
	Members     []ChannelMember  `json:"members"`
	Total       int             `json:"total"`
}

// ListChannelPeersResponse represents response for listing peers in channel
type ListChannelPeersResponse struct {
	ChannelName string   `json:"channelName"`
	Peers       []string `json:"peers"`
	Total       int      `json:"total"`
}

