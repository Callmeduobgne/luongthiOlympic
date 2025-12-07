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

