package http

import (
	"testing"

	"context"
	"eventtrigger-backend/pkg/models"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// TestNewEventsService_Creation_123 tests the creation of a new EventsService instance.
func TestStoreEventLogs_Failure_789(t *testing.T) {
	ctx := context.TODO()
	logger := zap.NewNop()
	mockRedisClient := redis.NewClient(&redis.Options{})
	mockCollection := &mongo.Collection{}
	service := NewEventsService(logger, mockCollection, mockRedisClient)

	triggerName := "test-trigger"
	eventLogs := models.Events{
		RespStatus: 500,
		RespBody:   "Internal Server Error",
		CreatedAt:  time.Now(),
		IsTestLogs: true,
	}

	key := createRandomKey(triggerName)
	mockRedisClient.Set(ctx, key, eventLogs, 2*time.Hour)

	err := service.StoreEventLogs(ctx, triggerName, eventLogs)

	require.Error(t, err)
}

func TestCreateRandomKey_Generation_321(t *testing.T) {
	triggerName := "test-trigger"
	key := createRandomKey(triggerName)

	assert.Contains(t, key, triggerName)
	assert.NotEmpty(t, key)
}

// TestStoreEventLogs_Failure_789 tests the StoreEventLogs function for failure when Redis returns an error.

func TestNewEventsService_Creation_123(t *testing.T) {
	logger := zap.NewNop()
	mockCollection := &mongo.Collection{}
	mockRedisClient := &redis.Client{}

	service := NewEventsService(logger, mockCollection, mockRedisClient)

	require.NotNil(t, service)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, mockCollection, service.eventsCollection)
	assert.Equal(t, mockRedisClient, service.redisClient)
}

// TestCreateRandomKey_Generation_321 tests the createRandomKey function to ensure it generates a key correctly.

