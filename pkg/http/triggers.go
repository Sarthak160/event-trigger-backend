package http

import (
	"context"
	"encoding/json"
	models "eventtrigger-backend/pkg/models"
	"eventtrigger-backend/pkg/services/cronjobs"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

type TriggerService struct {
	logger              *zap.Logger
	triggersCollection  *mongo.Collection
	cronJobSchedulerSvc *cronjobs.CronScheduler
	eventsSvc           *EventsService
}

func NewTriggerService(logger *zap.Logger, triggersCollection *mongo.Collection, cronJobSchedulerSvc *cronjobs.CronScheduler, eventsSvc *EventsService) *TriggerService {
	return &TriggerService{
		logger:              logger,
		triggersCollection:  triggersCollection,
		cronJobSchedulerSvc: cronJobSchedulerSvc,
		eventsSvc:           eventsSvc,
	}
}

func (es *TriggerService) TestTrigger(w http.ResponseWriter, r *http.Request) {
	var trigger models.Trigger
	if err := json.NewDecoder(r.Body).Decode(&trigger); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("trigger", trigger)
	callBackFunc, err := es.createHttpJob(trigger)

	if err != nil {
		es.logger.Error("error creating job", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	highOrderFuncToRemoveJobId := func() {
		callBackFunc()
		es.cronJobSchedulerSvc.RemoveJob(trigger.Name)
	}

	es.cronJobSchedulerSvc.AddJob(trigger.Name, trigger.Schedule, highOrderFuncToRemoveJobId)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("triggered registered for testing"))
}

func (es *TriggerService) CreateTrigger(w http.ResponseWriter, r *http.Request) {
	var trigger models.Trigger
	if err := json.NewDecoder(r.Body).Decode(&trigger); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := es.triggersCollection.InsertOne(r.Context(), trigger)

	if err != nil {
		es.logger.Error("error creating event", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	callBackFunc, err := es.createHttpJob(trigger)

	if err != nil {
		es.logger.Error("error creating job", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	es.cronJobSchedulerSvc.AddJob(trigger.Name, trigger.Schedule, callBackFunc)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("trigger created"))
}

func (es *TriggerService) DeleteTrigger(w http.ResponseWriter, r *http.Request) {
	eventName := chi.URLParam(r, "name")

	_, err := es.triggersCollection.DeleteOne(r.Context(), bson.M{"name": eventName})

	if err != nil {
		es.logger.Error("error deleting event", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	es.cronJobSchedulerSvc.RemoveJob(eventName)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("event deleted"))
}

func (es *TriggerService) ListTriggers(w http.ResponseWriter, r *http.Request) {
	var triggers []models.Trigger

	cursor, err := es.triggersCollection.Find(r.Context(), bson.M{})
	if err != nil {
		es.logger.Error("error listing triggers", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if err = cursor.All(r.Context(), &triggers); err != nil {
		es.logger.Error("error listing triggers", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	triggersBytes, err := json.Marshal(triggers)
	if err != nil {
		es.logger.Error("error listing triggers", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(triggersBytes)
}

func (es *TriggerService) GetTrigger(w http.ResponseWriter, r *http.Request) {
	eventName := chi.URLParam(r, "name")

	var trigger models.Trigger
	if err := es.triggersCollection.FindOne(r.Context(), bson.M{"name": eventName}).Decode(&trigger); err != nil {
		es.logger.Error("error getting trigger", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	triggerBytes, err := json.Marshal(trigger)
	if err != nil {
		es.logger.Error("error getting trigger", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(triggerBytes)
}

func (es *TriggerService) createHttpJob(trigger models.Trigger) (func(), error) {
	parsedURL, err := url.Parse(trigger.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// var payload interface{}
		// if err := json.Unmarshal([]byte(trigger.Payload), &payload); err != nil {
		// 	es.logger.Error("Failed to parse JSON payload",
		// 		zap.Error(err),
		// 	)
		// 	return
		// }

		// payloadBytes, err := json.Marshal(payload)
		// if err != nil {
		// 	es.logger.Error("Failed to marshal payload",
		// 		zap.Error(err),
		// 	)
		// 	return
		// }

		req, err := http.NewRequestWithContext(ctx, trigger.MethodType, parsedURL.String(), nil)
		if err != nil {
			es.logger.Error("Failed to create request",
				zap.Error(err),
				zap.String("endpoint", trigger.Endpoint),
			)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			es.logger.Error("Error executing scheduled request",
				zap.Error(err),
				zap.String("endpoint", trigger.Endpoint),
			)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		es.logger.Debug("Scheduled job response",
			zap.String("endpoint", trigger.Endpoint),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(body)),
		)
	}, nil
}
