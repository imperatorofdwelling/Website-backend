package metrics

import "fmt"

const (
	API_PROTOCOL = "https"
	API_VERSION  = 3
	API_ENDPOINT = "api.yookassa.ru"
)

const (
	PAYMENTS_ENDPOINT = "payments"
	PAYOUTS_ENDPOINT  = "payouts"
)

var (
	PAYMENTS_API = fmt.Sprintf("%v://%v/%v", API_PROTOCOL, API_VERSION, API_ENDPOINT)
)
