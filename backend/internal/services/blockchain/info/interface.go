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

import "context"

// InfoService is a common interface for blockchain info services
// This allows both Service (direct Fabric) and ServiceViaGateway to be used
type InfoService interface {
	GetBlockByNumber(ctx context.Context, blockNumber uint64) (*BlockInfo, error)
	GetChannelInfo(ctx context.Context) (*ChannelInfo, error)
	GetBlockByTxID(ctx context.Context, txID string) (*BlockInfo, error)
	GetTransactionByID(ctx context.Context, txID string) (string, error)
}

