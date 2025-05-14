package cronjobs

import (
	"sync"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type CronScheduler struct {
	cron   *cron.Cron
	jobs   map[string]cron.EntryID
	mu     sync.RWMutex
	logger *zap.Logger
}

func NewCronScheduler(logger *zap.Logger) *CronScheduler {
	return &CronScheduler{
		cron:   cron.New(cron.WithSeconds()),
		jobs:   make(map[string]cron.EntryID),
		logger: logger,
		mu:     sync.RWMutex{},
	}
}

func (cs *CronScheduler) AddJob(jobName string, schedule string, job func()) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	fmt.Println("Adding job", jobName, "with schedule", schedule)
	if existingID, exists := cs.jobs[jobName]; exists {
		cs.cron.Remove(existingID)
	}

	cs.logger.Debug("adding job", zap.String("name", jobName), zap.String("schedule", schedule))

	entryID, err := cs.cron.AddFunc(schedule, func() { job() })
	if err != nil {
		return err
	}

	cs.jobs[jobName] = entryID
	return nil
}

func (cs *CronScheduler) RemoveJob(jobName string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if entryID, exists := cs.jobs[jobName]; exists {
		cs.cron.Remove(entryID)
		delete(cs.jobs, jobName)
	}
}

func (cs *CronScheduler) Start() {
	cs.cron.Start()

	select {}
}

func (cs *CronScheduler) Stop() {
	cs.cron.Stop()
}
