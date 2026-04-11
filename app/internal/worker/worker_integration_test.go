package worker_test

import (
	integrationtest "backend/service-platform/app/internal/integrationtest"
	job "backend/service-platform/app/internal/job/constants"
	jobmanagers "backend/service-platform/app/internal/job/managers"
	"backend/service-platform/app/pkg/worker/handlers"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type WorkerSuite struct {
	integrationtest.RouterSuite
}

func TestWorkerSuite(t *testing.T) {
	suite.Run(t, new(WorkerSuite))
}

func (s *WorkerSuite) TestJobManagerAvailability() {
	// Test that JobManager is available in the test suite
	s.R.NotNil(s.Managers)
	s.R.NotNil(s.Managers.JobManager)
}

func (s *WorkerSuite) TestJobPriorities() {
	// Test that job priority constants are working correctly
	s.A.Equal("critical", job.PriorityCritical.String())
	s.A.Equal("high", job.PriorityHigh.String())
	s.A.Equal("normal", job.PriorityNormal.String())
	s.A.Equal("low", job.PriorityLow.String())
}

func (s *WorkerSuite) TestJobStatuses() {
	// Test that job status constants are working correctly
	s.A.Equal("pending", string(job.Pending))
	s.A.Equal("processing", string(job.Processing))
	s.A.Equal("completed", string(job.Completed))
	s.A.Equal("failed", string(job.Failed))
	s.A.Equal("retrying", string(job.Retrying))
}

func (s *WorkerSuite) TestCreateJobRequest() {
	// Test CreateJobRequest structure
	req := jobmanagers.CreateJobRequest{
		Type:        "test_job",
		Priority:    job.PriorityHigh,
		Payload:     map[string]interface{}{"test": "data"},
		MaxAttempts: 3,
	}

	s.A.Equal("test_job", req.Type)
	s.A.Equal(job.PriorityHigh, req.Priority)
	s.A.Equal("data", req.Payload["test"])
	s.A.Equal(3, req.MaxAttempts)
}

func (s *WorkerSuite) TestCreateAndRetrieveJob() {
	// Test creating and retrieving a job through JobManager
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	// Create a job
	req := jobmanagers.CreateJobRequest{
		Type:        "init_claim",
		Priority:    job.PriorityHigh,
		Payload:     map[string]interface{}{"user_id": "123", "amount": 100.0},
		MaxAttempts: 3,
	}

	createdJob, err := s.Managers.JobManager.CreateJob(ctx, req)
	s.R.NoError(err)
	s.R.NotNil(createdJob)
	s.A.Equal("init_claim", createdJob.Type)
	s.A.Equal(job.PriorityHigh, createdJob.Priority)
	s.A.Equal(job.Pending, createdJob.Status)

	// Retrieve the job by ID
	retrievedJob, err := s.Managers.JobManager.GetJob(ctx, createdJob.ID)
	s.R.NoError(err)
	s.R.NotNil(retrievedJob)
	s.A.Equal(createdJob.ID, retrievedJob.ID)
	s.A.Equal(createdJob.Type, retrievedJob.Type)
}

func (s *WorkerSuite) TestClaimHandler() {
	// Test claim handler directly
	claimHandler := handlers.NewInitClaimHandler(s.Resource.Logger)
	s.R.NotNil(claimHandler)

	ctx, cancel := context.WithTimeout(s.Ctx, 5*time.Second)
	defer cancel()

	// Create a claim job
	req := jobmanagers.CreateJobRequest{
		Type:     "init_claim",
		Priority: job.PriorityHigh,
		Payload: map[string]interface{}{
			"user_id":  "test-user-123",
			"amount":   250.75,
			"currency": "USD",
		},
		MaxAttempts: 3,
	}

	createdJob, err := s.Managers.JobManager.CreateJob(ctx, req)
	s.R.NoError(err)

	// Test handler execution
	err = claimHandler.Handle(ctx, createdJob)
	s.R.NoError(err, "Claim handler should process job successfully")
}

func (s *WorkerSuite) TestKYCHandler() {
	// Test KYC handler directly
	kycHandler := handlers.NewKYCVerificationHandler(s.Resource.Logger)
	s.R.NotNil(kycHandler)

	ctx, cancel := context.WithTimeout(s.Ctx, 5*time.Second)
	defer cancel()

	// Create a KYC job
	req := jobmanagers.CreateJobRequest{
		Type:     "kyc_verification",
		Priority: job.PriorityHigh,
		Payload: map[string]interface{}{
			"user_id":       "test-user-456",
			"document_id":   "doc-123",
			"document_type": "passport",
		},
		MaxAttempts: 3,
	}

	createdJob, err := s.Managers.JobManager.CreateJob(ctx, req)
	s.R.NoError(err)

	// Test handler execution
	err = kycHandler.Handle(ctx, createdJob)
	s.R.NoError(err, "KYC handler should process job successfully")
}

func (s *WorkerSuite) TestJobsByStatus() {
	// Test retrieving jobs by status
	ctx, cancel := context.WithTimeout(s.Ctx, 10*time.Second)
	defer cancel()

	// Create multiple jobs with different statuses
	jobs := []jobmanagers.CreateJobRequest{
		{
			Type:        "init_claim",
			Priority:    job.PriorityHigh,
			Payload:     map[string]interface{}{"user_id": "user1"},
			MaxAttempts: 3,
		},
		{
			Type:        "kyc_verification",
			Priority:    job.PriorityNormal,
			Payload:     map[string]interface{}{"user_id": "user2"},
			MaxAttempts: 3,
		},
	}

	var createdJobs []string
	for _, jobReq := range jobs {
		createdJob, err := s.Managers.JobManager.CreateJob(ctx, jobReq)
		s.R.NoError(err)
		createdJobs = append(createdJobs, createdJob.ID.String())
	}

	// Retrieve pending jobs
	pendingJobs, err := s.Managers.JobManager.GetJobsByStatus(ctx, job.Pending, 10)
	s.R.NoError(err)
	s.R.GreaterOrEqual(len(pendingJobs), 2, "Should have at least the jobs we created")

	// Verify our created jobs are in the pending list
	foundJobs := 0
	for _, pendingJob := range pendingJobs {
		for _, createdJobID := range createdJobs {
			if pendingJob.ID.String() == createdJobID {
				foundJobs++
				s.A.Equal(job.Pending, pendingJob.Status)
			}
		}
	}
	s.A.Equal(2, foundJobs, "Should find both created jobs in pending status")
}
