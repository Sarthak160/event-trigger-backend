package http

import (
	"testing"

	"context"
	"eventtrigger-backend/pkg/models"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap/zaptest"
)

// TestCreateRandomKey_Success_789 tests the successful generation of a random key.
func TestStoreEventLogs_RedisSetError_301(t *testing.T) {
	logger := zaptest.NewLogger(t)
	s, err := miniredis.Run()
	require.NoError(t, err)
	// s.Close() // Close immediately to simulate connection error

	redisClient := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	// No defer redisClient.Close() as we will close miniredis server first.

	es := NewEventsService(logger, nil, redisClient)

	// Close miniredis server to cause Set to fail
	s.Close()
	// Now any operation on redisClient should fail

	ctx := context.TODO()
	triggerName := "failing-trigger"
	eventLogs := models.Events{
		RespStatus: 500,
		RespBody:   "Event failure",
		CreatedAt:  time.Now(),
		IsTestLogs: true,
	}

	storeErr := es.StoreEventLogs(ctx, triggerName, eventLogs)
	require.Error(t, storeErr, "Expected an error when Redis Set fails")
	// The exact error message might vary depending on the redis client and OS,
	// so checking for a non-nil error is usually sufficient.
	// Example: "dial tcp 127.0.0.1:XXXX: connect: connection refused" or "redis: client is closed"
	// For more specific check: assert.Contains(t, storeErr.Error(), "connection refused") or similar

	// Note: The logger output (es.logger.Error) can be captured and verified if needed,
	// but zaptest primarily logs to test output.
	redisClient.Close() // Clean up client resources
}

func TestCreateRandomKey_Success_789(t *testing.T) {
	triggerName := "test-trigger"
	key := createRandomKey(triggerName)
	require.NotEmpty(t, key)
	require.Contains(t, key, triggerName)
}

// TestNewEventsService_Nominal_101 tests the instantiation of EventsService.

func TestNewEventsService_Nominal_101(t *testing.T) {
	logger := zaptest.NewLogger(t)
	var mockEventsCollection *mongo.Collection // Can be nil or empty struct as it's not used by StoreEventLogs

	// For redisClient, we can use a nil client for this specific test if we only check fields,
	// but for consistency, using a miniredis instance is also fine.
	// Here, we'll just check if the fields are assigned.
	// In actual usage tests for methods, a real (mocked) client is needed.
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"}) // Dummy client for instantiation test
	defer rdb.Close()

	es := NewEventsService(logger, mockEventsCollection, rdb)

	require.NotNil(t, es)
	assert.Equal(t, logger, es.logger)
	assert.Equal(t, mockEventsCollection, es.eventsCollection)
	assert.Equal(t, rdb, es.redisClient)
}

// TestStoreEventLogs_RedisSetError_301 simulates a Redis failure (e.g., server down)
// and ensures that the error is propagated and logged correctly.

