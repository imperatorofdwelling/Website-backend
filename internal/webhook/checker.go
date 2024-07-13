package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/imperatorofdwelling/Website-backend/internal/metrics"
	"github.com/imperatorofdwelling/Website-backend/pkg/repository/redis"

	"github.com/google/uuid"
)

var httpClientDo = func(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

type WebhookData struct {
	YooKassaTransactionID string                       `json:"yooKassaTransactionID"`
	YouKassaConfirmation  metrics.YouKassaConfirmation `json:"youKassaConfirmation"`
	ServerUUID            uuid.UUID                    `json:"serverUUID"`
}

type Checker struct {
	redisDB redis.RedisInterface
}

func NewChecker(redisDB redis.RedisInterface) *Checker {
	return &Checker{redisDB: redisDB}
}

func NewWebhookData(yooKassaID string, serverUUID uuid.UUID) *WebhookData {
	return &WebhookData{
		YooKassaTransactionID: yooKassaID,
		ServerUUID:            serverUUID,
	}
}

var (
	NotNeedToCheck     = errors.New("not need to check")
	CannotStartToCheck = errors.New("can't start checking, redis is empty")
	EmptyResponse      = errors.New("empty response")
)

func (c *Checker) StartCheck(whData *WebhookData, startStatus metrics.Status) error {
	if startStatus.IsAlreadyProcessedStatus() {
		return NotNeedToCheck
	}
	err := c.redisDB.CommitTransaction(whData.ServerUUID, startStatus)
	if err != nil {
		return err
	}
	go c.updater(whData)
	return nil
}

func (c *Checker) updater(whData *WebhookData) {
	ch := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	go c.signaller(ch, ctx)
	for range ch {
		newStatus, _ := c.sendCheckRequstToYouKassa(whData)
		err := c.updateRedis(whData, newStatus)
		if c.isFinalUpdate(newStatus, err) {
			break
		}
	}
	cancel()
}

func (c *Checker) updateRedis(whData *WebhookData, r *CheckResponse) error {
	if r == nil {
		return EmptyResponse
	}
	exists := c.redisDB.ExistsTransaction(whData.ServerUUID)
	if !exists {
		return CannotStartToCheck
	}
	return c.redisDB.CommitTransaction(whData.ServerUUID, r.Status)
}

func (c *Checker) isFinalUpdate(r *CheckResponse, err error) bool {
	if err != nil {
		return false
	}
	return r.Status.IsAlreadyProcessedStatus()
}

func (c *Checker) signaller(ch chan<- struct{}, ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		fibArr := c.getFibArr()
		for _, timing := range fibArr {
			sleepTiming := time.Duration(timing) * time.Minute
			c.sleep(sleepTiming, ctx)
			ch <- struct{}{}
		}
		close(ch)
	}
}

func (c *Checker) sleep(d time.Duration, ctx context.Context) {
	timer := time.NewTimer(d)
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		return
	}
}

func (c *Checker) getFibArr() []int {
	var fibArr = []int{1, 1}
	var fibSum int
	maxFibSum := metrics.CheckMaxMinutes
	index := 2
	for {
		nextFibNum := fibArr[index-1] + fibArr[index-2]
		if fibSum+nextFibNum >= maxFibSum {
			return fibArr
		}
		fibArr = append(fibArr, nextFibNum)
		index++
	}
}

type CheckResponse struct {
	Status metrics.Status `json:"status"`
}

func (c *Checker) sendCheckRequstToYouKassa(whData *WebhookData) (*CheckResponse, error) {
	url := metrics.PaymentsApi + metrics.PayoutsEndpoint + "/" + whData.YooKassaTransactionID
	apiReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	apiReq.SetBasicAuth(metrics.GetConfirmationData())
	apiReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClientDo(apiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var statusResponse *CheckResponse
	err = json.Unmarshal(respBody, &statusResponse)
	if err != nil {
		return nil, err
	}
	return statusResponse, nil
}
