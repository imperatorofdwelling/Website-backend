package tests

import (
	"encoding/json"
	"fmt"
	"github.com/https-whoyan/dwellingPayload/internal/metrics"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPayment(t *testing.T) {
	req, err := http.NewRequest("POST", "/payment/create", nil)

	defer req.Body.Close()

	resp := new(metrics.PaymentResponse)
	json.NewDecoder(req.Body).Decode(resp)
	fmt.Println(resp)

	assert.Error(t, err)
}
