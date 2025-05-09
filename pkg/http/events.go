package http

import (
	"context"
	"crypto/rand"
	"eventtrigger-backend/pkg/models"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type EventsService struct {
	logger           *zap.Logger
	eventsCollection *mongo.Collection
	redisClient      *redis.Client
}

func NewEventsService(logger *zap.Logger, eventsCollection *mongo.Collection, redisClient *redis.Client) *EventsService {
	return &EventsService{
		logger:           logger,
		eventsCollection: eventsCollection,
		redisClient:      redisClient,
	}
}

func (es *EventsService) StoreEventLogs(ctx context.Context, triggerName string, eventLogs models.Events) error {
	key := createRandomKey(triggerName)
	es.logger.Debug("storing event logs", zap.String("key", key))
	err := es.redisClient.Set(ctx, key, eventLogs, 2*time.Hour).Err()
	if err != nil {
		es.logger.Error("error storing event logs", zap.Error(err))
		return err
	}

	return nil
}

func createRandomKey(triggerName string) string {

	b := make([]byte, 8)

	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s-%x", triggerName, b)
}
