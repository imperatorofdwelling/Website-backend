package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/https-whoyan/dwellingPayload/internal/models"
	myJson "github.com/https-whoyan/dwellingPayload/pkg/json"
	"github.com/https-whoyan/dwellingPayload/pkg/repository/postgres"
	"io"
	"log/slog"
	"net/http"
	"time"
	"unicode"
)

// _______________________
// Save Refillable Card
// _______________________

// SaveCard accepted structure from frontend
type SaveCard struct {
	UserId   string `json:"user_id"`
	Synonym  string `json:"synonym"`
	FirstSix string `json:"first_six"`
	LastFour string `json:"last_four"`
}

// SaveCardResponse response to frontend
type SaveCardResponse struct {
	Status Status `json:"status"`
	Error  string `json:"error"`
}

func (c SaveCard) isFullData() bool {
	if c.UserId == "" {
		return false
	}
	if c.Synonym == "" {
		return false
	}
	if len(c.FirstSix) != 6 || len(c.LastFour) != 4 {
		return false
	}
	firstSixIsContainsOfDigits := isNumeric(c.FirstSix)
	lastFourIsContainsOfDigits := isNumeric(c.LastFour)

	return firstSixIsContainsOfDigits && lastFourIsContainsOfDigits
}

func SaveCardHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const fn = "endpoints.SaveCardHandler"

		log = log.With(slog.String("fn", fn))
		log.Debug("safe card endpoint called")

		c := new(SaveCard)

		if err := myJson.Read(r, c); err != nil {
			log.Error("failed to read request", slog.String("error", err.Error()))
			myJson.Write(w, http.StatusBadRequest, NewErrorResponse("bad request"))
			return
		}
		if !c.isFullData() {
			myJson.Write(w, http.StatusBadRequest, NewErrorResponse("bad request, not full data"))
			return
		}

		usedUUID, _ := uuid.Parse(c.UserId)
		insertedUsed := models.User{
			Id:      usedUUID,
			Balance: 0,
		}
		cardMask, _ := models.GenerateCardMask(c.FirstSix, c.LastFour)
		insertedCard := models.NewRefillableCard(&insertedUsed, c.Synonym, cardMask)

		db, isContains := postgres.GetDB()
		if !isContains {
			log.Error("failed to get database")
			myJson.Write(w, http.StatusInternalServerError, NewErrorResponse(
				"Internal server error, database isn't initialized"))
			return
		}
		isUpdated, err := db.InsertOrUpdateRefillableCard(insertedCard)
		if err != nil {
			log.Error("failed to insert or update refillable card", slog.String("error", err.Error()))
			myJson.Write(w, http.StatusInternalServerError, NewErrorResponse(
				"Internal server error!"))
			return
		}

		var messageToLog string
		if isUpdated {
			messageToLog = fmt.Sprintf("Card info updated successfully, userID: %v", c.UserId)
		} else {
			messageToLog = fmt.Sprintf("Card info insert successfully, userID: %v", c.UserId)
		}

		log.Info(messageToLog)

		// Create JSON Response
		backendResponse := SaveCardResponse{
			Status: "success",
			Error:  "",
		}

		// Send response to Frontend
		myJson.Write(w, http.StatusOK, backendResponse)

		log.Info("response to frontend successfully sent")

	}
}

// ________________
// Payout
// ________________

const (
	DefaultDescription = "From ImperatorOfDwelling for renting an apartment."
)

// PayoutRequestEndpoint endpoint parameters
type PayoutRequestEndpoint struct {
	ToUserId string `json:"user_id"`
	Amount   Amount `json:"amount"`
}

func (p PayoutRequestEndpoint) isFullData() bool {
	if p.ToUserId == "" {
		return false
	}
	if p.Amount.Value == "" || p.Amount.Currency == "" {
		return false
	}
	return true
}

// PayloadRequestKassa provided json paraments of payout request
// https://yookassa.ru/developers/payouts/making-payouts/bank-card/using-payout-widget/making-payouts-with-synonym
type PayloadRequestKassa struct {
	Amount      Amount `json:"amount"`
	CardSynonym string `json:"card_synonym"`
	Description string `json:"description"`
}

// YooKassaPayloadModel YooKassa payload model
type YooKassaPayloadModel struct {
	Amount      Amount    `json:"amount"`
	Status      Status    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Test        bool      `json:"test"`
}

func Payload(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const fn = "endpoints.Payload"

		log = log.With(slog.String("fn", fn))
		log.Debug("payload endpoint called")

		req := new(PayoutRequestEndpoint)

		if err := myJson.Read(r, req); err != nil {
			log.Error("failed to read request", slog.String("error", err.Error()))
			myJson.Write(w, http.StatusBadRequest, NewErrorResponse("bad request"))
			return
		}

		if !req.isFullData() {
			myJson.Write(w, http.StatusBadRequest, NewErrorResponse("provided not full data"))
			return
		}

		currDB, exists := postgres.GetDB()
		if !exists {
			log.Error("failed to get database")
			myJson.Write(w, http.StatusInternalServerError,
				NewErrorResponse("internal server error, database isn't initialized"),
			)
			return
		}

		uuidUser, err := uuid.Parse(req.ToUserId)
		if err != nil {
			log.Error("failed to parse userID from request", slog.String("error", err.Error()))
			myJson.Write(w, http.StatusBadRequest, NewErrorResponse("bad request"))
			return
		}
		row, err := currDB.GetRefillableCardByUserID(uuidUser)
		cardSynonym := row.CardSynonym
		if err != nil {
			log.Error("failed to get database refillable card", slog.String("error", err.Error()))
			myJson.Write(w, http.StatusInternalServerError, NewErrorResponse(
				"internal server error, error getting mask card"),
			)
			return
		}
		if cardSynonym == "" {
			log.Info("failed to get database refillable card", slog.String("error", err.Error()))
			myJson.Write(w, http.StatusLocked, NewErrorResponse("the card is untethered for userID"))
			return
		}

		createReq := createPayloadBody(req, cardSynonym)

		resp, err := sendPayloadRequest(createReq)
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

		paymentResp := new(YooKassaPayloadModel)

		// Create JSON Response
		if err := json.Unmarshal(respBody, paymentResp); err != nil {
			log.Error(
				"failed to make json from response",
				slog.String("error", err.Error()),
			)
			myJson.Write(w, http.StatusInternalServerError, NewErrorResponse("server error"))
			return
		}

		// Send response to Frontend
		myJson.Write(w, http.StatusOK, paymentResp)

		log.Info("response to frontend successfully sent")

	}
}

// toUserID := create.ToUserId
//	db, isContains := repository.GetDB()

func createPayloadBody(c *PayoutRequestEndpoint, cardSynonym string) *PayloadRequestKassa {
	createReq := &PayloadRequestKassa{
		Amount:      c.Amount,
		CardSynonym: cardSynonym,
		Description: DefaultDescription,
	}
	return createReq
}

func sendPayloadRequest(r *PayloadRequestKassa) (*http.Response, error) {
	createReqJson, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	apiReq, err := http.NewRequest(
		"POST",
		PAYMENTS_API+PAYOUTS_ENDPOINT,
		bytes.NewBuffer(createReqJson),
	)
	if err != nil {
		return nil, err
	}
	idempotenceKey := uuid.New().String()

	apiReq.SetBasicAuth(storeID, secretKey)
	apiReq.Header.Set("Idempotence-Key", idempotenceKey)
	apiReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(apiReq)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ______________
// Utils
// _______________

func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
