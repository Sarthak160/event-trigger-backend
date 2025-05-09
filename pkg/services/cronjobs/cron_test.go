package cronjobs

import (
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestNewCronScheduler_Creation_001 tests the creation of a new CronScheduler instance.
func TestAddJob_NewJob_002(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewCronScheduler(logger)

	jobName := "test-job"
	schedule := "* * * * * *"
	job := func() {}

	err := scheduler.AddJob(jobName, schedule, job)
	require.NoError(t, err)

	scheduler.mu.RLock()
	defer scheduler.mu.RUnlock()
	_, exists := scheduler.jobs[jobName]
	assert.True(t, exists)
}

// TestAddJob_InvalidSchedule_003 tests adding a job with an invalid schedule to the CronScheduler.

func TestRemoveJob_ExistingJob_004(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewCronScheduler(logger)

	jobName := "test-job"
	schedule := "* * * * * *"
	job := func() {}

	err := scheduler.AddJob(jobName, schedule, job)
	require.NoError(t, err)

	scheduler.RemoveJob(jobName)

	scheduler.mu.RLock()
	defer scheduler.mu.RUnlock()
	_, exists := scheduler.jobs[jobName]
	assert.False(t, exists)
}

// TestStartScheduler_006 tests starting the CronScheduler.

func TestStartScheduler_006(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewCronScheduler(logger)

	go func() {
		scheduler.Start()
	}()

	// Allow some time for the scheduler to start
	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, scheduler.cron)
}

// TestStopScheduler_007 tests stopping the CronScheduler.

func TestStopScheduler_007(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewCronScheduler(logger)

	go func() {
		scheduler.Start()
	}()

	// Allow some time for the scheduler to start
	time.Sleep(100 * time.Millisecond)

	scheduler.Stop()

	assert.NotNil(t, scheduler.cron)
}

func TestNewCronScheduler_Creation_001(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewCronScheduler(logger)

	require.NotNil(t, scheduler)
	assert.NotNil(t, scheduler.cron)
	assert.NotNil(t, scheduler.jobs)
	assert.NotNil(t, scheduler.logger)
	assert.NotNil(t, scheduler.mu)
}

// TestAddJob_NewJob_002 tests adding a new job to the CronScheduler.

func TestAddJob_InvalidSchedule_003(t *testing.T) {
	logger := zap.NewNop()
	scheduler := NewCronScheduler(logger)

	jobName := "test-job"
	schedule := "invalid-schedule"
	job := func() {}

	err := scheduler.AddJob(jobName, schedule, job)
	require.Error(t, err)
}

// TestRemoveJob_ExistingJob_004 tests removing an existing job from the CronScheduler.

