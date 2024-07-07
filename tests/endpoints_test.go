package tests

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/https-whoyan/dwellingPayload/config"
	"github.com/https-whoyan/dwellingPayload/internal/metrics"
	srv "github.com/https-whoyan/dwellingPayload/internal/server/http"
	logger2 "github.com/https-whoyan/dwellingPayload/pkg/logger"
	"github.com/https-whoyan/dwellingPayload/pkg/repository"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	dbCfg  = config.LoadConfig("../.env").PostgresSQLConfig
	logger = logger2.New(logger2.EnvLocal)
	router http.Handler
)

type testCase struct {
	name           string
	requestBody    *metrics.Create
	expectedStatus int
	expectedError  string
}

func Init() {
	if err := repository.InitPostgresDB(dbCfg); err != nil {
		logger.Error("failed to init DB instance", slog.String("error", err.Error()))
	}
	db, _ := repository.GetDB()
	logRepo := repository.NewLogRepository(db)
	router = srv.NewRouter(logger, logRepo)
}

func TestPayment(t *testing.T) {
	Init()
	testCases := []testCase{
		{
			name:           "OK",
			requestBody:    metrics.NewCreate(uuid.New().String(), "100.00", "RUB"),
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Bad request empty body",
			requestBody:    &metrics.Create{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "userId or amount is empty",
		},
		{
			name:           "Bad request invalid currency",
			requestBody:    metrics.NewCreate(uuid.New().String(), "345.5", "USD"),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Incorrect currency of payment. The value of the amount.currency parameter doesn't correspond with the settings of your store. Specify another currency value in the request or contact the YooMoney manager to change the settings",
		},
		{
			name:           "Bad request invalid value",
			requestBody:    metrics.NewCreate(uuid.New().String(), "-123.43", "RUB"),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Error in the payment amount. Specify the amount in correct format. For example, 100.00",
		},
		{
			name: "Bad request invalid id",
			requestBody: &metrics.Create{
				Amount: metrics.Amount{
					Currency: "RUB",
					Value:    "450.5",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "userId or amount is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reqBodyBytes, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/payment/create", bytes.NewBuffer(reqBodyBytes))

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedError != "" {
				respBody := new(metrics.ErrorResponse)
				_ = json.NewDecoder(rr.Body).Decode(respBody)
				assert.Equal(t, tc.expectedError, respBody.Error)
			}
		})
	}
}
