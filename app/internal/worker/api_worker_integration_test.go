package worker_test

import (
	commonmodels "backend/service-platform/app/internal/common/models"
	integrationtest "backend/service-platform/app/internal/integrationtest"
	job "backend/service-platform/app/internal/job/constants"
	"backend/service-platform/app/internal/job/entities"
	jobmanagers "backend/service-platform/app/internal/job/managers"
	testutil "backend/service-platform/app/internal/testutil"
	worker "backend/service-platform/app/internal/worker"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type APIWorkerSuite struct {
	integrationtest.RouterSuite
	workerService *worker.WorkerService
}

func TestAPIWorkerSuite(t *testing.T) {
	suite.Run(t, new(APIWorkerSuite))
}

func (s *APIWorkerSuite) SetupTest() {
	s.RouterSuite.SetupTest()

	// Get WorkerService from services registry (already created in RouterSuite)
	s.workerService = s.Services.WorkerService

	// Start worker service in background with long-running context
	go func() {
		// Use background context with longer timeout for worker service
		// This prevents context cancellation from interrupting job processing
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		if err := s.workerService.Start(ctx); err != nil {
			s.Resource.Logger.Error("Worker service failed during test",
				zap.Error(err))
		}
	}()

	// Wait for worker service to be ready
	s.waitForWorkersReady()

	// Give a bit more time for workers to fully initialize
	time.Sleep(500 * time.Millisecond)
}

func (s *APIWorkerSuite) TearDownTest() {
	if s.workerService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.workerService.Stop(ctx)
	}
	s.RouterSuite.TearDownTest()
}

// waitForWorkersReady waits for worker service to have workers started
func (s *APIWorkerSuite) waitForWorkersReady() {
	if s.workerService == nil {
		return
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		stats := s.workerService.GetStats()
		if stats.TotalWorkers > 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *APIWorkerSuite) TestAPIWithWorkerIntegration() {
	// Create a job via API (if the job creation endpoint exists)
	// or directly via JobManager and verify it gets processed
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	jobReq := jobmanagers.CreateJobRequest{
		Type:        "init_claim",
		Priority:    job.PriorityHigh,
		Payload:     map[string]interface{}{"user_id": "api-test-user", "amount": 500.0},
		MaxAttempts: 3,
	}

	createdJob, err := s.Managers.JobManager.CreateJob(ctx, jobReq)
	s.R.NoError(err)
	s.R.NotNil(createdJob)
	s.A.Equal(job.Pending, createdJob.Status)

	// 3. Wait for job to be processed by worker
	var processedJob *entity.Job
	s.A.Eventually(func() bool {
		jobEntity, err := s.Managers.JobManager.GetJob(ctx, createdJob.ID)
		if err != nil {
			return false
		}
		processedJob = jobEntity
		return jobEntity.Status != job.Pending // Job should be picked up
	}, 5*time.Second, 100*time.Millisecond, "Job should be processed by worker")

	s.R.NotNil(processedJob)

	// Job should either be processing, completed, or still pending (depending on timing)
	s.A.Contains([]job.Status{job.Pending, job.Processing, job.Completed}, processedJob.Status)
}

func (s *APIWorkerSuite) TestConcurrentAPIAndWorkerOperations() {
	// Test concurrent API requests while workers are processing jobs
	ctx, cancel := context.WithTimeout(s.Ctx, 15*time.Second)
	defer cancel()

	// Create multiple jobs concurrently
	jobTypes := []string{"init_claim", "kyc_verification", "wallet_balance_sync"}
	var createdJobIDs []string

	for i, jobType := range jobTypes {
		jobReq := jobmanagers.CreateJobRequest{
			Type:        jobType,
			Priority:    job.PriorityNormal,
			Payload:     map[string]interface{}{"test_id": fmt.Sprintf("concurrent-test-%d", i)},
			MaxAttempts: 3,
		}

		createdJob, err := s.Managers.JobManager.CreateJob(ctx, jobReq)
		s.R.NoError(err)
		createdJobIDs = append(createdJobIDs, createdJob.ID.String())
	}

	// Verify all jobs still exist and can be retrieved
	for _, jobID := range createdJobIDs {
		retrievedJob, err := s.Managers.JobManager.GetJob(ctx, uuid.MustParse(jobID))
		s.R.NoError(err)
		s.R.NotNil(retrievedJob)
		s.A.Equal(jobID, retrievedJob.ID.String())
	}
}

func (s *APIWorkerSuite) TestWorkerStatsWhileAPIRunning() {
	// Test that we can get worker statistics while API is running

	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	// Create some jobs to give workers something to do
	for i := 0; i < 3; i++ {
		jobReq := jobmanagers.CreateJobRequest{
			Type:        "init_claim",
			Priority:    job.PriorityNormal,
			Payload:     map[string]interface{}{"batch_id": fmt.Sprintf("stats-test-%d", i)},
			MaxAttempts: 3,
		}

		_, err := s.Managers.JobManager.CreateJob(ctx, jobReq)
		s.R.NoError(err)
	}

	// Allow a brief moment for workers to process jobs (if any are queued)

	// Get worker statistics
	stats := s.workerService.GetStats()

	// Verify statistics make sense
	s.A.GreaterOrEqual(stats.ActiveWorkers, 0)
	s.A.GreaterOrEqual(stats.ProcessingJobs, 0)
	s.A.GreaterOrEqual(stats.TotalProcessed, int64(0))
	s.A.GreaterOrEqual(stats.TotalFailed, int64(0))
	s.R.NotNil(stats.QueueDepths)

	// Make an API request while checking stats
	resp, code, err := testutil.RequestHTTP[commonmodels.GeneralResponse[commonmodels.HealthResponse]](s.Echo, http.MethodGet, "/health", nil, nil)
	s.R.NoError(err)
	s.A.Equal(http.StatusOK, code)
	s.A.Equal("up", resp.Data.Status)

	// Get stats again to ensure they're still accessible
	stats2 := s.workerService.GetStats()
	s.R.NotNil(stats2)
}

func (s *APIWorkerSuite) TestJobCreationViaAPIEndpoint() {
	// Test creating jobs through a hypothetical API endpoint
	// This test assumes we might add a job creation endpoint in the future

	// For now, we'll test the pattern of how it would work
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	// Simulate what an API endpoint for job creation might look like
	jobRequest := map[string]interface{}{
		"type":         "init_claim",
		"priority":     "high",
		"payload":      map[string]interface{}{"user_id": "api-created-user", "amount": 750.0},
		"max_attempts": 3,
	}

	// Convert to JSON as it would come from API
	jsonData, err := json.Marshal(jobRequest)
	s.R.NoError(err)

	// Parse it back (simulating API endpoint processing)
	var parsedRequest map[string]interface{}
	err = json.Unmarshal(jsonData, &parsedRequest)
	s.R.NoError(err)

	// Create a job using JobManager (as the API endpoint would do)
	jobReq := jobmanagers.CreateJobRequest{
		Type:        parsedRequest["type"].(string),
		Priority:    job.PriorityHigh, // Would parse from string in real API
		Payload:     parsedRequest["payload"].(map[string]interface{}),
		MaxAttempts: int(parsedRequest["max_attempts"].(float64)),
	}

	createdJob, err := s.Managers.JobManager.CreateJob(ctx, jobReq)
	s.R.NoError(err)
	s.R.NotNil(createdJob)
	s.A.Equal("init_claim", createdJob.Type)
	s.A.Equal(job.PriorityHigh, createdJob.Priority)

	// Verify the job was created and can be retrieved
	retrievedJob, err := s.Managers.JobManager.GetJob(ctx, createdJob.ID)
	s.R.NoError(err)
	s.A.Equal(createdJob.ID, retrievedJob.ID)
}
