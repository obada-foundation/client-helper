package blockchain

import (
	node "github.com/obada-foundation/client-helper/system/obadanode"
	"go.uber.org/zap"
)

// Service holds service dependencies.
type Service struct {
	nodeClient node.Client
	logger     *zap.SugaredLogger
}

// NewService creates a new instance of the service.
func NewService(client node.Client, logger *zap.SugaredLogger) *Service {
	return &Service{
		nodeClient: client,
		logger:     logger,
	}
}
