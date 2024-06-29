package metrics

import (
	"fmt"
	"os"
)

// ___________________
// YooKassa protocol
// ___________________

const (
	ApiProtocol = "https"
	ApiVersion  = 3
	ApiEndpoint = "api.yookassa.ru"
)

const (
	PaymentsEndpoint = "payments"
	PayoutsEndpoint  = "payouts"
)

var (
	PaymentsApi = fmt.Sprintf("%v://%v/v%v/", ApiProtocol, ApiEndpoint, ApiVersion)
)

type Status string

const (
	Succeeded         Status = "succeeded"
	Canceled          Status = "canceled"
	WaitingForCapture Status = "waiting_for_capture"
	Pending           Status = "pending"
)

func (status Status) IsAlreadyProcessedStatus() bool {
	switch status {
	case Succeeded, Canceled:
		return true
	}
	return false
}

// _______________________
// YouKassa confirmation
// _______________________

type YouKassaConfirmation struct {
	StoreID        string
	StoreSecretKey string
}

var (
	confirmationInstance YouKassaConfirmation
)

func Init() {
	confirmationInstance = YouKassaConfirmation{
		StoreID:        os.Getenv("STORE_ID"),
		StoreSecretKey: os.Getenv("SECRET_KEY"),
	}
}

func GetConfirmationData() (storeID string, storeSecretKey string) {
	return confirmationInstance.StoreID, confirmationInstance.StoreSecretKey
}

// An indication of how many minutes I have to check the status
const (
	CheckMaxMinutes = 24 * 60
)
