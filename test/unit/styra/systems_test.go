package styra_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bankdata/styra-controller/pkg/styra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) request(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	args := m.Called(ctx, method, url, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

// ClientInterface is an interface that has the same methods as your Client.
type ClientInterface interface {
	request(ctx context.Context, method, url string, body io.Reader) (*http.Response, error)
}

func TestGetSystemByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "/v1/systems?name=test")
		// Send response to be tested
		rw.Write([]byte(`{"result": []}`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Use Client & URL from our local test server
	c := &styra.Client{HTTPClient: *server.Client(), URL: server.URL}
	ctx := context.Background()

	resp, err := c.GetSystemByName(ctx, "test")

	// Assert no error
	assert.NoError(t, err)

	// Assert status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Assert SystemConfig
	assert.Nil(t, resp.SystemConfig)
}

func TestGetSystemByNameOneExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "/v1/systems?name=test")
		// Send response to be tested
		rw.Write([]byte(`{"result": [
			{
				"name": "test", 
				"id": "abc123"
			}
		]}`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Use Client & URL from our local test server
	c := &styra.Client{HTTPClient: *server.Client(), URL: server.URL}
	ctx := context.Background()

	resp, err := c.GetSystemByName(ctx, "test")

	// Assert no error
	assert.NoError(t, err)

	// Assert status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Assert SystemConfig
	assert.Equal(t, "test", resp.SystemConfig.Name)
	assert.Equal(t, "abc123", resp.SystemConfig.ID)
}

func TestGetSystemByNameMoreExist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "/v1/systems?name=test")
		// Send response to be tested
		rw.Write([]byte(`{"result": [
			{
				"name": "test", 
				"id": "abc123"
			},
			
			{
				"name": "test", 
				"id": "def456"
			}
		]}`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Use Client & URL from our local test server
	c := &styra.Client{HTTPClient: *server.Client(), URL: server.URL}
	ctx := context.Background()

	resp, err := c.GetSystemByName(ctx, "test")

	// Assert no error
	assert.NoError(t, err)

	// Assert status code
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Assert SystemConfig
	assert.Equal(t, "test", resp.SystemConfig.Name)
	assert.Equal(t, "abc123", resp.SystemConfig.ID)
}

