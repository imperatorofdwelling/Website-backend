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
			expectedStatus: http.StatusBadRequest,
			expectedError:  "bad request, not full data",
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
			expectedError:  "bad request, not full data",
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
		newTc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reqBodyBytes, _ := json.Marshal(newTc.requestBody)
			req, _ := http.NewRequest("POST", "/save_card", bytes.NewBuffer(reqBodyBytes))

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
