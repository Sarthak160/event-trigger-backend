package http

import (
	"eventtrigger-backend/pkg/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestCreateHttpJob_InvalidURL_018 tests createHttpJob when trigger.Endpoint is invalid.
func TestCreateHttpJob_Callback_RequestCreationError_019(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ts := &TriggerService{logger: logger}

	trigger := models.Trigger{Endpoint: "http://example.com", MethodType: "INVALID METHOD\r\n"}

	jobFunc, err := ts.createHttpJob(trigger)
	require.NoError(t, err)
	require.NotNil(t, jobFunc)

	// This will log "Failed to create request"
	// No panic expected, function should return gracefully.
	assert.NotPanics(t, jobFunc)
}

// TestCreateHttpJob_Callback_ClientDoError_020 tests callback failure on client.Do (network error).

func TestCreateHttpJob_Callback_ClientDoError_020(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ts := &TriggerService{logger: logger}

	trigger := models.Trigger{Endpoint: "http://localhost:1", MethodType: http.MethodGet} // Port 1 is usually inaccessible

	jobFunc, err := ts.createHttpJob(trigger)
	require.NoError(t, err)
	require.NotNil(t, jobFunc)

	// This will log "Error executing scheduled request"
	assert.NotPanics(t, jobFunc)
}

func TestCreateHttpJob_InvalidURL_018(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ts := &TriggerService{logger: logger}

	trigger := models.Trigger{Endpoint: "http://[::1]:namedport"}

	jobFunc, err := ts.createHttpJob(trigger)

	assert.Error(t, err)
	assert.Nil(t, jobFunc)
	assert.Contains(t, err.Error(), "invalid URL")
}

// TestCreateHttpJob_Callback_RequestCreationError_019 tests callback failure on request creation (invalid HTTP method).

