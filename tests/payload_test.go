package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imperatorofdwelling/Website-backend/internal/endpoints"
	"github.com/stretchr/testify/assert"
)

type payloadTest struct {
	name           string
	requestBody    *endpoints.PayoutRequestEndpoint
	expectedStatus int
	expectedError  string
}

func TestPayload(t *testing.T) {
	Init()
	testCases := []payloadTest{
		{
			name: "OK",
			requestBody: &endpoints.PayoutRequestEndpoint{
				ToUserId: "69c1f84f-8fd8-480b-b5fe-4aaf96826791",
				Amount: endpoints.Amount{
					Currency: "RUB",
					Value:    "100",
				},
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Bad request empty body",
			requestBody:    &endpoints.PayoutRequestEndpoint{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "provided not full data",
		},
		{
			name: "Bad request invalid currency",
			requestBody: &endpoints.PayoutRequestEndpoint{
				ToUserId: "",
				Amount: endpoints.Amount{
					Currency: "AKJ",
					Value:    "100",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "provided not full data",
		},
		{
			name: "User not have a card",
			requestBody: &endpoints.PayoutRequestEndpoint{
				ToUserId: "&&&",
				Amount: endpoints.Amount{
					Currency: "RUB",
					Value:    "1827.98",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad request",
		},
		{
			name: "Bad request invalid id",
			requestBody: &endpoints.PayoutRequestEndpoint{
				ToUserId: "&&&",
				Amount: endpoints.Amount{
					Currency: "RUB",
					Value:    "1827.98",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad request",
		},
	}

	for _, tc := range testCases {
		newTc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reqBodyBytes, _ := json.Marshal(newTc.requestBody)
			req, _ := http.NewRequest("POST", "/payload/create", bytes.NewBuffer(reqBodyBytes))

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, newTc.expectedStatus, rr.Code)

			if newTc.expectedError != "" {
				respBody := new(endpoints.ErrorResponse)
				_ = json.NewDecoder(rr.Body).Decode(respBody)
				assert.Equal(t, newTc.expectedError, respBody.Error)
			}
		})
	}
}
