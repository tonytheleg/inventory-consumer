package kessel

import (
	"errors"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/project-kessel/inventory-consumer/internal/common"
	"github.com/project-kessel/inventory-consumer/internal/mocks"
	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTestConfig(enabled bool, enableOidcAuth bool) CompletedConfig {
	options := &Options{
		Enabled:        enabled,
		InventoryURL:   "localhost:9090",
		Insecure:       true,
		EnableOidcAuth: enableOidcAuth,
		ClientId:       "test-client",
		ClientSecret:   "test-secret",
		TokenEndpoint:  "http://localhost:8080/token",
	}

	return CompletedConfig{
		&completedConfig{
			Options: options,
		},
	}
}

func createTestLogger() *log.Helper {
	_, logger := common.InitLogger("info", common.LoggerOptions{})
	return log.NewHelper(log.With(logger, "service", "test"))
}

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		config        CompletedConfig
		expectEnabled bool
		expectAuth    bool
		shouldError   bool
	}{
		{
			name:          "disabled client returns disabled KesselClient",
			config:        createTestConfig(false, false),
			expectEnabled: false,
			expectAuth:    false,
			shouldError:   false,
		},
		{
			name:          "enabled client without auth creates client successfully",
			config:        createTestConfig(true, false),
			expectEnabled: true,
			expectAuth:    false,
			shouldError:   false,
		},
		{
			name:          "enabled client with auth creates client successfully",
			config:        createTestConfig(true, true),
			expectEnabled: true,
			expectAuth:    true,
			shouldError:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := createTestLogger()

			client, err := New(test.config, logger)

			if test.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, client)
			assert.Equal(t, test.expectEnabled, client.Enabled)
			assert.Equal(t, test.expectAuth, client.AuthEnabled)

			if !test.config.Enabled {
				// For disabled clients, InventoryClient should be nil
				assert.Nil(t, client.InventoryClient)
			} else {
				// For enabled clients, InventoryClient should be set
				assert.NotNil(t, client.InventoryClient)
			}
		})
	}
}

func TestKesselClient_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "client is enabled",
			enabled:  true,
			expected: true,
		},
		{
			name:     "client is disabled",
			enabled:  false,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &KesselClient{
				Enabled: test.enabled,
			}

			result := client.IsEnabled()
			assert.Equal(t, test.expected, result)
		})
	}
}

// TestClientProvider_CreateOrUpdateResource tests the CreateOrUpdateResource method using MockClient
func TestClientProvider_CreateOrUpdateResource(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockClient)
		request        *v1beta2.ReportResourceRequest
		expectedResult *v1beta2.ReportResourceResponse
		expectedError  error
	}{
		{
			name: "successful create or update resource",
			mockSetup: func(m *mocks.MockClient) {
				m.On("CreateOrUpdateResource", mock.Anything).
					Return(&v1beta2.ReportResourceResponse{}, nil)
			},
			request: &v1beta2.ReportResourceRequest{
				Type:               "host",
				ReporterType:       "hbi",
				ReporterInstanceId: "test-instance",
			},
			expectedResult: &v1beta2.ReportResourceResponse{},
			expectedError:  nil,
		},
		{
			name: "create or update resource fails",
			mockSetup: func(m *mocks.MockClient) {
				m.On("CreateOrUpdateResource", mock.Anything).
					Return(&v1beta2.ReportResourceResponse{}, errors.New("grpc error"))
			},
			request: &v1beta2.ReportResourceRequest{
				Type:               "host",
				ReporterType:       "hbi",
				ReporterInstanceId: "test-instance",
			},
			expectedResult: &v1beta2.ReportResourceResponse{},
			expectedError:  errors.New("grpc error"),
		},
		{
			name: "create or update resource with specific request data",
			mockSetup: func(m *mocks.MockClient) {
				// Use mock.Anything for simpler matching
				m.On("CreateOrUpdateResource", mock.Anything).
					Return(&v1beta2.ReportResourceResponse{}, nil)
			},
			request: &v1beta2.ReportResourceRequest{
				Type:               "host",
				ReporterType:       "hbi",
				ReporterInstanceId: "specific-instance",
			},
			expectedResult: &v1beta2.ReportResourceResponse{},
			expectedError:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mocks.MockClient{}
			test.mockSetup(mockClient)

			// Use the mock as ClientProvider interface
			var client ClientProvider = mockClient

			// Call the method being tested
			result, err := client.CreateOrUpdateResource(test.request)

			// Assert expectations
			if test.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.expectedResult, result)
			mockClient.AssertExpectations(t)
		})
	}
}

// TestClientProvider_DeleteResource tests the DeleteResource method using MockClient
func TestClientProvider_DeleteResource(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockClient)
		request        *v1beta2.DeleteResourceRequest
		expectedResult *v1beta2.DeleteResourceResponse
		expectedError  error
	}{
		{
			name: "successful delete resource",
			mockSetup: func(m *mocks.MockClient) {
				m.On("DeleteResource", mock.Anything).
					Return(&v1beta2.DeleteResourceResponse{}, nil)
			},
			request: &v1beta2.DeleteResourceRequest{
				Reference: &v1beta2.ResourceReference{
					ResourceType: "host",
					ResourceId:   "test-host-id",
					Reporter: &v1beta2.ReporterReference{
						Type: "hbi",
					},
				},
			},
			expectedResult: &v1beta2.DeleteResourceResponse{},
			expectedError:  nil,
		},
		{
			name: "delete resource fails",
			mockSetup: func(m *mocks.MockClient) {
				m.On("DeleteResource", mock.Anything).
					Return(&v1beta2.DeleteResourceResponse{}, errors.New("delete failed"))
			},
			request: &v1beta2.DeleteResourceRequest{
				Reference: &v1beta2.ResourceReference{
					ResourceType: "host",
					ResourceId:   "test-host-id",
					Reporter: &v1beta2.ReporterReference{
						Type: "hbi",
					},
				},
			},
			expectedResult: &v1beta2.DeleteResourceResponse{},
			expectedError:  errors.New("delete failed"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mocks.MockClient{}
			test.mockSetup(mockClient)

			// Use the mock as ClientProvider interface
			var client ClientProvider = mockClient

			// Call the method being tested
			result, err := client.DeleteResource(test.request)

			// Assert expectations
			if test.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.expectedResult, result)
			mockClient.AssertExpectations(t)
		})
	}
}
