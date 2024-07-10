package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/imperatorofdwelling/Website-backend/internal/metrics"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	TransactionTable = "transactionTable"
	expiration       = time.Minute * metrics.CheckMaxMinutes
)

var (
	ctx = context.Background()
)

func getKey(serverTransactionID uuid.UUID) string {
	return TransactionTable + ":" + serverTransactionID.String()
}

// ___________
// Saving
// ___________

var (
	TransactionAlreadyExistsError = fmt.Errorf("transactionAlreadyExistsError")
	TransactionNotFoundError      = fmt.Errorf("transactionNotFoundError")
	ChangedKeyErr                 = fmt.Errorf("the key changed at the time of the request")
)

func (r *RedisDB) CommitTransaction(serverTransactionID uuid.UUID, status metrics.Status) error {
	if r.ExistsTransaction(serverTransactionID) {
		return TransactionAlreadyExistsError
	}

	pipe := r.rdb.TxPipeline()
	pipe.Set(ctx, getKey(serverTransactionID), status, expiration)
	_, err := pipe.Exec(ctx)
	return err
}

// Updater

func (r *RedisDB) UpdateStatus(serverTransactionID uuid.UUID, status metrics.Status) error {
	if !r.ExistsTransaction(serverTransactionID) {
		return TransactionNotFoundError
	}

	lifeSpan, err := r.rdb.TTL(ctx, getKey(serverTransactionID)).Result()
	if err != nil {
		return err
	}
	duration := lifeSpan * time.Second
	pipe := r.rdb.TxPipeline()
	pipe.Set(ctx, getKey(serverTransactionID), status, duration)
	_, err = pipe.Exec(ctx)
	return err
}

// __________
// Getters
// __________

func (r *RedisDB) GetTransactionStatus(serverTransactionID uuid.UUID, status metrics.Status) (metrics.Status, error) {
	if !r.ExistsTransaction(serverTransactionID) {
		return "", TransactionNotFoundError
	}

	checkedKey := getKey(serverTransactionID)
	resultStatus := metrics.Status("")

	err := r.rdb.Watch(ctx, func(tx *redis.Tx) error {
		transactionStatus, err := tx.Get(ctx, checkedKey).Result()
		if err != nil {
			return err
		}
		resultStatus = metrics.Status(transactionStatus)
		return nil
	}, checkedKey)
	if errors.Is(err, redis.TxFailedErr) {
		return "", ChangedKeyErr
	}

	if resultStatus.IsAlreadyProcessedStatus() {
		defer r.DelKey(serverTransactionID)
	}
	return resultStatus, nil
}

func (r *RedisDB) ExistsTransaction(serverTransactionID uuid.UUID) bool {
	val, err := r.rdb.Exists(ctx, getKey(serverTransactionID)).Result()
	return err == nil && val == 1
}

// Deleter

func (r *RedisDB) DelKey(serverTransactionID uuid.UUID) { r.rdb.Del(ctx, getKey(serverTransactionID)) }
