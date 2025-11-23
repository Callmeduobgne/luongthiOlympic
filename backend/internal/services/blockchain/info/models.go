// Copyright 2024 IBN Network (ICTU Blockchain Network)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

