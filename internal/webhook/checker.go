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

// _______________
// Checking
// _______________

// as webhook

type WebhookData struct {
	YooKassaTransactionID string                       `json:"yooKassaTransactionID"`
	YouKassaConfirmation  metrics.YouKassaConfirmation `json:"youKassaConfirmation"`
	ServerUUID            uuid.UUID                    `json:"serverUUID"`
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

// StartCheck starts periodically checking the status of transaction
func StartCheck(whData *WebhookData, startStatus metrics.Status) error {
	if startStatus.IsAlreadyProcessedStatus() {
		return NotNeedToCheck
	}
	currDB, contains := redis.GetCurrRedisDB()
	if !contains {
		return CannotStartToCheck
	}
	err := currDB.CommitTransaction(whData.ServerUUID, startStatus)
	if err != nil {
		return err
	}
	go updater(whData)

	return nil
}

func updater(whData *WebhookData) {
	ch := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	go signaller(ch, ctx)
	for range ch {
		newStatus, _ := sendCheckRequstToYouKassa(whData)
		err := updateRedis(whData, newStatus)
		if isFinalUpdate(newStatus, err) {
			break
		}
	}
	cancel()
}

func updateRedis(whData *WebhookData, r *CheckResponse) error {
	currRedis, contains := redis.GetCurrRedisDB()
	if !contains {
		return CannotStartToCheck
	}
	if r == nil {
		return EmptyResponse
	}

	return currRedis.CommitTransaction(whData.ServerUUID, r.Status)
}

func isFinalUpdate(r *CheckResponse, err error) bool {
	if err != nil {
		return false
	}
	return r.Status.IsAlreadyProcessedStatus()
}

// Code for signaling that a request should be made

func signaller(ch chan<- struct{}, ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		fibArr := getFibArr()

		for _, timing := range fibArr {
			sleepTiming := time.Duration(timing) * time.Minute
			sleep(sleepTiming, ctx)
			ch <- struct{}{}
		}
		close(ch)
	}

}

func sleep(d time.Duration, ctx context.Context) {
	timer := time.NewTimer(d)

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		return
	}
}

func getFibArr() []int {
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

// _____________________
// Request to youKassa
// _____________________

type CheckResponse struct {
	Status metrics.Status `json:"status"`
}

func sendCheckRequstToYouKassa(whData *WebhookData) (*CheckResponse, error) {
	url := metrics.PaymentsApi + metrics.PayoutsEndpoint + "/" + whData.YooKassaTransactionID
	apiReq, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	apiReq.SetBasicAuth(metrics.GetConfirmationData())
	apiReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(apiReq)

	if err != nil {
		return nil, err
	}

	// Read Response
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
