package webhook

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/imperatorofdwelling/Website-backend/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Redis DB
type MockRedisDB struct {
	mock.Mock
}

func (m *MockRedisDB) CommitTransaction(serverUUID uuid.UUID, status metrics.Status) error {
	args := m.Called(serverUUID, status)
	return args.Error(0)
}

func (m *MockRedisDB) ExistsTransaction(serverUUID uuid.UUID) bool {
	args := m.Called(serverUUID)
	return args.Bool(0)
}

func (m *MockRedisDB) UpdateStatus(serverUUID uuid.UUID, status metrics.Status) error {
	args := m.Called(serverUUID, status)
	return args.Error(0)
}

func (m *MockRedisDB) GetTransactionStatus(serverUUID uuid.UUID, status metrics.Status) (metrics.Status, error) {
	args := m.Called(serverUUID, status)
	return args.Get(0).(metrics.Status), args.Error(1)
}

func (m *MockRedisDB) DelKey(serverUUID uuid.UUID) {
	m.Called(serverUUID)
}

func TestUpdateRedis(t *testing.T) {
	mockDB := new(MockRedisDB)
	checker := NewChecker(mockDB)

	testCases := []struct {
		name         string
		whData       *WebhookData
		response     *CheckResponse
		existsReturn bool
		mockReturn   error
		expectedErr  error
	}{
		{
			name:         "Successful update",
			whData:       NewWebhookData("123", uuid.New()),
			response:     &CheckResponse{Status: metrics.Succeeded},
			existsReturn: true,
			mockReturn:   nil,
			expectedErr:  nil,
		},
		{
			name:         "Redis commit error",
			whData:       NewWebhookData("123", uuid.New()),
			response:     &CheckResponse{Status: metrics.Succeeded},
			existsReturn: true,
			mockReturn:   errors.New("redis error"),
			expectedErr:  errors.New("redis error"),
		},
		{
			name:        "Empty response",
			whData:      NewWebhookData("123", uuid.New()),
			response:    nil,
			expectedErr: EmptyResponse,
		},
		{
			name:         "Transaction does not exist",
			whData:       NewWebhookData("123", uuid.New()),
			response:     &CheckResponse{Status: metrics.Succeeded},
			existsReturn: false,
			expectedErr:  CannotStartToCheck,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.response != nil {
				mockDB.On("ExistsTransaction", tc.whData.ServerUUID).Return(tc.existsReturn)
				if tc.existsReturn {
					mockDB.On("CommitTransaction", tc.whData.ServerUUID, tc.response.Status).Return(tc.mockReturn)
				}
			}

			err := checker.updateRedis(tc.whData, tc.response)

			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestIsFinalUpdate(t *testing.T) {
	checker := NewChecker(nil) // We don't need a real Redis for this test

	testCases := []struct {
		name     string
		response *CheckResponse
		err      error
		expected bool
	}{
		{
			name:     "Final status",
			response: &CheckResponse{Status: metrics.Succeeded},
			err:      nil,
			expected: true,
		},
		{
			name:     "Non-final status",
			response: &CheckResponse{Status: metrics.Pending},
			err:      nil,
			expected: false,
		},
		{
			name:     "Error occurred",
			response: nil,
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := checker.isFinalUpdate(tc.response, tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetFibArr(t *testing.T) {
	checker := NewChecker(nil) // We don't need a real Redis for this test
	result := checker.getFibArr()
	expected := []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377, 610, 987}
	assert.Equal(t, expected, result)
}

func TestSignaller(t *testing.T) {
	// Создаем функцию, имитирующую signaller с контролируемыми входными данными
	testSignaller := func(fibArr []int, sleepDuration time.Duration) (int, time.Duration) {
		ch := make(chan struct{}, len(fibArr))
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		start := time.Now()
		go func() {
			for range fibArr {
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(sleepDuration) // Имитация sleep
					ch <- struct{}{}
				}
			}
			close(ch)
		}()

		count := 0
		for range ch {
			count++
		}
		duration := time.Since(start)

		return count, duration
	}

	testCases := []struct {
		name           string
		fibArr         []int
		sleepDuration  time.Duration
		expectedCount  int
		expectedMinDur time.Duration
	}{
		{
			name:           "Short sequence",
			fibArr:         []int{1, 1, 2},
			sleepDuration:  10 * time.Millisecond,
			expectedCount:  3,
			expectedMinDur: 30 * time.Millisecond,
		},
		{
			name:           "Longer sequence",
			fibArr:         []int{1, 1, 2, 3, 5},
			sleepDuration:  5 * time.Millisecond,
			expectedCount:  5,
			expectedMinDur: 25 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, duration := testSignaller(tc.fibArr, tc.sleepDuration)
			assert.Equal(t, tc.expectedCount, count, "Unexpected number of signals")
			assert.True(t, duration >= tc.expectedMinDur, "Duration was shorter than expected")
		})
	}
}

func TestSleep(t *testing.T) {
	checker := NewChecker(nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	start := time.Now()
	checker.sleep(100*time.Millisecond, ctx)
	duration := time.Since(start)

	assert.True(t, duration >= 100*time.Millisecond, "Sleep duration was shorter than expected")
	assert.True(t, duration < 150*time.Millisecond, "Sleep duration was longer than expected")

	// Тест отмены контекста
	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start = time.Now()
	checker.sleep(1*time.Second, ctx)
	duration = time.Since(start)

	assert.True(t, duration < 100*time.Millisecond, "Sleep did not cancel quickly enough")
}

func TestStartCheck(t *testing.T) {
	mockDB := new(MockRedisDB)
	checker := NewChecker(mockDB)

	testCases := []struct {
		name        string
		whData      *WebhookData
		startStatus metrics.Status
		commitError error
		expectedErr error
	}{
		{
			name:        "Already processed status",
			whData:      NewWebhookData("123", uuid.New()),
			startStatus: metrics.Succeeded,
			expectedErr: NotNeedToCheck,
		},
		{
			name:        "Commit transaction error",
			whData:      NewWebhookData("123", uuid.New()),
			startStatus: metrics.Pending,
			commitError: errors.New("commit error"),
			expectedErr: errors.New("commit error"),
		},
		{
			name:        "Successful start",
			whData:      NewWebhookData("123", uuid.New()),
			startStatus: metrics.Pending,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.commitError != nil {
				mockDB.On("CommitTransaction", tc.whData.ServerUUID, tc.startStatus).Return(tc.commitError)
			} else if tc.expectedErr == nil {
				mockDB.On("CommitTransaction", tc.whData.ServerUUID, tc.startStatus).Return(nil)
			}

			err := checker.StartCheck(tc.whData, tc.startStatus)

			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestSendCheckRequstToYouKassa(t *testing.T) {
	httpClientDoOriginal := httpClientDo
	defer func() {
		httpClientDo = httpClientDoOriginal
	}()

	checker := NewChecker(nil)

	testCases := []struct {
		name           string
		whData         *WebhookData
		httpDoFunc     func(req *http.Request) (*http.Response, error)
		expectedStatus *CheckResponse
		expectedErr    string
	}{
		{
			name: "Error creating request",
			whData: &WebhookData{
				YooKassaTransactionID: "",
			},
			httpDoFunc: func(req *http.Request) (*http.Response, error) {
				// Simулируем ошибку создания запроса, возвращая ошибку парсинга URL.
				return nil, errors.New("parse \"http://:invalid-url\": invalid URI for request")
			},
			expectedStatus: nil,
			expectedErr:    "invalid URI for request",
		},
		{
			name: "Error executing request",
			whData: &WebhookData{
				YooKassaTransactionID: "123",
			},
			httpDoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("http error")
			},
			expectedStatus: nil,
			expectedErr:    "http error",
		},
		{
			name: "Error reading response body",
			whData: &WebhookData{
				YooKassaTransactionID: "123",
			},
			httpDoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(&errReader{}),
				}, nil
			},
			expectedStatus: nil,
			expectedErr:    "test error",
		},
		{
			name: "Error unmarshaling response",
			whData: &WebhookData{
				YooKassaTransactionID: "123",
			},
			httpDoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("invalid json")),
				}, nil
			},
			expectedStatus: nil,
			expectedErr:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "Successful response",
			whData: &WebhookData{
				YooKassaTransactionID: "123",
			},
			httpDoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"status":"succeeded"}`)),
				}, nil
			},
			expectedStatus: &CheckResponse{Status: metrics.Succeeded},
			expectedErr:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.httpDoFunc != nil {
				httpClientDo = tc.httpDoFunc
			}

			statusResponse, err := checker.sendCheckRequstToYouKassa(tc.whData)

			if tc.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedStatus, statusResponse)
			}
		})
	}
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}
func (errReader) Close() error {
	return nil
}
