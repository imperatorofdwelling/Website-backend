package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/imperatorofdwelling/Website-backend/internal/endpoints"
	"github.com/stretchr/testify/assert"
)

type saveCardTestCase struct {
	name           string
	requestBody    *endpoints.SaveCard
	expectedStatus int
	expectedError  string
}

func TestSaveCard(t *testing.T) {
	Init()
	testCases := []saveCardTestCase{
		{
			name: "OK",
			requestBody: &endpoints.SaveCard{
				UserId:   uuid.New().String(),
				Synonym:  "testSinonim1",
				FirstSix: "000000",
				LastFour: "9999",
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Bad request empty body",
			requestBody:    &endpoints.SaveCard{},
			expectedStatus: http.StatusNotFound,
			expectedError:  "userId or amount is empty",
		},
		{
			name: "Bad synonym",
			requestBody: &endpoints.SaveCard{
				UserId:   uuid.New().String(),
				Synonym:  "",
				FirstSix: "000000",
				LastFour: "9999",
			},

			expectedStatus: http.StatusBadRequest,
			expectedError:  "Incorrect currency of payment. The value of the amount.currency parameter doesn't correspond with the settings of your store. Specify another currency value in the request or contact the YooMoney manager to change the settings",
		},
		{
			name: "Bad request incorrect digits 1",
			requestBody: &endpoints.SaveCard{
				UserId:   uuid.New().String(),
				Synonym:  "sdmhsdk",
				FirstSix: "sjdhs",
				LastFour: "9999",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad request, not full data",
		},
		{
			name: "Bad request incorrect digits 2",
			requestBody: &endpoints.SaveCard{
				UserId:   uuid.New().String(),
				Synonym:  "sdfkjsdfsdkf",
				FirstSix: "1212",
				LastFour: "898",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad request, not full data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reqBodyBytes, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/save_card", bytes.NewBuffer(reqBodyBytes))

			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedError != "" {
				respBody := new(endpoints.ErrorResponse)
				_ = json.NewDecoder(rr.Body).Decode(respBody)
				assert.Equal(t, tc.expectedError, respBody.Error)
			}
		})
	}
}
