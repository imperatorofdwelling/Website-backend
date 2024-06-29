package endpoints

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/imperatorofdwelling/Website-backend/internal/metrics"
	"github.com/imperatorofdwelling/Website-backend/internal/webhook"
	myJson "github.com/imperatorofdwelling/Website-backend/pkg/json"
	"github.com/imperatorofdwelling/Website-backend/pkg/repository/postgres"

	"github.com/google/uuid"
)

// Request from frontend

type Create struct {
	UserId string `json:"user_id"`
	Amount Amount `json:"amount"`
}

// request to YoooKassa API

type CreatePaymentRequest struct {
	Amount       Amount       `json:"amount"`
	Confirmation Confirmation `json:"confirmation"`
	Capture      bool         `json:"capture"`
	Description  string       `json:"description"`
}

// PaymentResponse response from youkassa
type PaymentResponse struct {
	ID           string               `json:"id"`
	Status       metrics.Status       `json:"status"`
	Paid         bool                 `json:"paid"`
	Amount       Amount               `json:"amount"`
	Confirmation ConfirmationResponse `json:"confirmation"`
	CreatedAt    string               `json:"created_at"`
	Description  string               `json:"description"`
	Metadata     interface{}          `json:"metadata"`
	Recipient    Recipient            `json:"recipient"`
	Refundable   bool                 `json:"refundable"`
	Test         bool                 `json:"test"`
}

// PaymentAnswer response (answer) to frontend
type PaymentAnswer struct {
	TransactionId uuid.UUID        `json:"transaction_id"`
	Status        metrics.Status   `json:"status"`
	YouKassaModel *PaymentResponse `json:"you_kassa_model"`
}

func NewPaymentAnswer(youKassaModel *PaymentResponse) *PaymentAnswer {
	newUUID, _ := uuid.NewUUID()
	return &PaymentAnswer{
		TransactionId: newUUID,
		Status:        metrics.Pending,
		YouKassaModel: youKassaModel,
	}
}

// Payment amount

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type Confirmation struct {
	Type string `json:"type"`
}

type Recipient struct {
	AccountID string `json:"account_id"`
	GatewayID string `json:"gateway_id"`
}

type ConfirmationResponse struct {
	Type              string `json:"type"`
	ConfirmationToken string `json:"confirmation_token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(message string) *ErrorResponse {
	return &ErrorResponse{
		Error: message,
	}
}

type PaymentHandler struct {
	log       *slog.Logger
	logWriter postgres.LogRepository
}

func NewPaymentHandler(log *slog.Logger, db postgres.LogRepository) *PaymentHandler {
	return &PaymentHandler{
		log:       log,
		logWriter: db,
	}
}

func (h *PaymentHandler) Payment(w http.ResponseWriter, r *http.Request) {
	const fn = "endpoints.Payment"

	log := h.log.With(slog.String("fn", fn))
	log.Debug("payment endpoint called")

	req := new(Create)

	if err := myJson.Read(r, req); err != nil {
		log.Error("failed to read request", slog.String("error", err.Error()))
		myJson.Write(w, http.StatusBadRequest, NewErrorResponse("bad request"))
		return
	}

	createReq := createPaymentBody(req)

	resp, err := sendRequest(createReq)
	if err != nil {
		log.Error(
			"failed to send request to YooKassa API",
			slog.String("error", err.Error()),
		)
		myJson.Write(w, http.StatusInternalServerError, NewErrorResponse("server error"))
		return
	}

	log.Debug("request to API sent")

	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(
			"failed to read response",
			slog.String("error", err.Error()),
		)
		myJson.Write(w, http.StatusInternalServerError, NewErrorResponse("server error"))
		return
	}

	paymentRespFromYoukassa := new(PaymentResponse)

	// Create JSON Response
	if err := json.Unmarshal(respBody, paymentRespFromYoukassa); err != nil {
		log.Error(
			"failed to make json from response",
			slog.String("error", err.Error()),
		)
		myJson.Write(w, http.StatusInternalServerError, NewErrorResponse("server error"))
		return
	}

	paymentResp := NewPaymentAnswer(paymentRespFromYoukassa)
	checkerData := webhook.NewWebhookData(paymentResp.YouKassaModel.ID, paymentResp.TransactionId)
	_ = webhook.StartCheck(checkerData, paymentResp.Status)

	log.Info("response to frontend successfully sent")

	logToDb := postgres.NewLog(paymentRespFromYoukassa.ID, req.Amount.Value, string(paymentRespFromYoukassa.Status))

	err = h.logWriter.InsertLog(logToDb)
	if err != nil {
		log.Error("failed to write log to db", slog.String("error", err.Error()))
		myJson.Write(w, http.StatusInternalServerError, NewErrorResponse("server error"))
		return
	}
	log.Info("log to db successfully written")

	// Send response to Frontend
	myJson.Write(w, http.StatusOK, paymentRespFromYoukassa)
}

func createPaymentBody(create *Create) *CreatePaymentRequest {
	orderNum := uuid.New().String()
	createReq := &CreatePaymentRequest{
		Amount: create.Amount,
		Confirmation: Confirmation{
			Type: "embedded",
		},
		Capture:     true,
		Description: "Заказ № " + orderNum,
	}
	return createReq
}

func sendRequest(createReq *CreatePaymentRequest) (*http.Response, error) {
	createReqJson, err := json.Marshal(createReq)
	if err != nil {
		return nil, err
	}
	apiReq, err := http.NewRequest(
		"POST",
		metrics.PaymentsApi+metrics.PaymentsEndpoint,
		bytes.NewBuffer(createReqJson),
	)
	if err != nil {
		return nil, err
	}
	idempotenceKey := uuid.New().String()

	apiReq.SetBasicAuth(metrics.GetConfirmationData())
	apiReq.Header.Set("Idempotence-Key", idempotenceKey)
	apiReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(apiReq)

	if err != nil {
		return nil, err
	}

	return resp, nil
}
