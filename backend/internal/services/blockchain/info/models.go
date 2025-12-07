package info

// BlockInfo represents simplified block information (no protobuf parsing)
type BlockInfo struct {
	BlockNumber uint64 `json:"block_number"`
	RawBlock    string `json:"raw_block_hex"` // Hex-encoded block data
	Size        int    `json:"size_bytes"`
}

// ChannelInfo represents channel information
type ChannelInfo struct {
	ChannelID string `json:"channel_id"`
	RawInfo   string `json:"raw_info_hex"` // Hex-encoded blockchain info
	Size      int    `json:"size_bytes"`
}

// BlockQueryRequest for flexible block queries
type BlockQueryRequest struct {
	BlockNumber *uint64 `json:"block_number,omitempty"`
	TxID        *string `json:"tx_id,omitempty"`
}

