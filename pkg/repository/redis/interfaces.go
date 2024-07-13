package redis

import (
	"github.com/google/uuid"
	"github.com/imperatorofdwelling/Website-backend/internal/metrics"
)

//go:generate mockery --name RedisInterface
type RedisInterface interface {
	CommitTransaction(serverTransactionID uuid.UUID, status metrics.Status) error
	UpdateStatus(serverTransactionID uuid.UUID, status metrics.Status) error
	GetTransactionStatus(serverTransactionID uuid.UUID, status metrics.Status) (metrics.Status, error)
	ExistsTransaction(serverTransactionID uuid.UUID) bool
	DelKey(serverTransactionID uuid.UUID)
}
