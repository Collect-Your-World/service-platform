package integration

import (
	"backend/service-platform/app/database/constant/job"
	"backend/service-platform/app/manager"
	"backend/service-platform/app/pkg/worker/handlers"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type WorkerSuite struct {
	RouterSuite
}

func TestWorkerSuite(t *testing.T) {
	suite.Run(t, new(WorkerSuite))
}

func (s *WorkerSuite) TestJobManagerAvailability() {
	// Test that JobManager is available in the test suite
	s.r.NotNil(s.managers)
	s.r.NotNil(s.managers.JobManager)
}

func (s *WorkerSuite) TestJobPriorities() {
	// Test that job priority constants are working correctly
	s.a.Equal("critical", job.PriorityCritical.String())
	s.a.Equal("high", job.PriorityHigh.String())
	s.a.Equal("normal", job.PriorityNormal.String())
	s.a.Equal("low", job.PriorityLow.String())
}

func (s *WorkerSuite) TestJobStatuses() {
	// Test that job status constants are working correctly
	s.a.Equal("pending", string(job.Pending))
	s.a.Equal("processing", string(job.Processing))
	s.a.Equal("completed", string(job.Completed))
	s.a.Equal("failed", string(job.Failed))
	s.a.Equal("retrying", string(job.Retrying))
}

func (s *WorkerSuite) TestCreateJobRequest() {
	// Test CreateJobRequest structure
	req := manager.CreateJobRequest{
		Type:        "test_job",
		Priority:    job.PriorityHigh,
		Payload:     map[string]interface{}{"test": "data"},
		MaxAttempts: 3,
	}

	s.a.Equal("test_job", req.Type)
	s.a.Equal(job.PriorityHigh, req.Priority)
	s.a.Equal("data", req.Payload["test"])
	s.a.Equal(3, req.MaxAttempts)
}

func (s *WorkerSuite) TestCreateAndRetrieveJob() {
	// Test creating and retrieving a job through JobManager
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	// Create a job
	req := manager.CreateJobRequest{
		Type:        "init_claim",
		Priority:    job.PriorityHigh,
		Payload:     map[string]interface{}{"user_id": "123", "amount": 100.0},
		MaxAttempts: 3,
	}

	createdJob, err := s.managers.JobManager.CreateJob(ctx, req)
	s.r.NoError(err)
	s.r.NotNil(createdJob)
	s.a.Equal("init_claim", createdJob.Type)
	s.a.Equal(job.PriorityHigh, createdJob.Priority)
	s.a.Equal(job.Pending, createdJob.Status)

	// Retrieve the job by ID
	retrievedJob, err := s.managers.JobManager.GetJob(ctx, createdJob.ID)
	s.r.NoError(err)
	s.r.NotNil(retrievedJob)
	s.a.Equal(createdJob.ID, retrievedJob.ID)
	s.a.Equal(createdJob.Type, retrievedJob.Type)
}

func (s *WorkerSuite) TestClaimHandler() {
	// Test claim handler directly
	claimHandler := handlers.NewInitClaimHandler(s.resource.Logger)
	s.r.NotNil(claimHandler)

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	// Create a claim job
	req := manager.CreateJobRequest{
		Type:     "init_claim",
		Priority: job.PriorityHigh,
		Payload: map[string]interface{}{
			"user_id":  "test-user-123",
			"amount":   250.75,
			"currency": "USD",
		},
		MaxAttempts: 3,
	}

	createdJob, err := s.managers.JobManager.CreateJob(ctx, req)
	s.r.NoError(err)

	// Test handler execution
	err = claimHandler.Handle(ctx, createdJob)
	s.r.NoError(err, "Claim handler should process job successfully")
}

func (s *WorkerSuite) TestKYCHandler() {
	// Test KYC handler directly
	kycHandler := handlers.NewKYCVerificationHandler(s.resource.Logger)
	s.r.NotNil(kycHandler)

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	// Create a KYC job
	req := manager.CreateJobRequest{
		Type:     "kyc_verification",
		Priority: job.PriorityHigh,
		Payload: map[string]interface{}{
			"user_id":       "test-user-456",
			"document_id":   "doc-123",
			"document_type": "passport",
		},
		MaxAttempts: 3,
	}

	createdJob, err := s.managers.JobManager.CreateJob(ctx, req)
	s.r.NoError(err)

	// Test handler execution
	err = kycHandler.Handle(ctx, createdJob)
	s.r.NoError(err, "KYC handler should process job successfully")
}

func (s *WorkerSuite) TestWalletHandler() {
	// Wallet handler not implemented yet
	s.T().Skip("Wallet handler not implemented")
}

func (s *WorkerSuite) TestJobsByStatus() {
	// Test retrieving jobs by status
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	// Create multiple jobs with different statuses
	jobs := []manager.CreateJobRequest{
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
		createdJob, err := s.managers.JobManager.CreateJob(ctx, jobReq)
		s.r.NoError(err)
		createdJobs = append(createdJobs, createdJob.ID)
	}

	// Retrieve pending jobs
	pendingJobs, err := s.managers.JobManager.GetJobsByStatus(ctx, job.Pending, 10)
	s.r.NoError(err)
	s.r.GreaterOrEqual(len(pendingJobs), 2, "Should have at least the jobs we created")

	// Verify our created jobs are in the pending list
	foundJobs := 0
	for _, pendingJob := range pendingJobs {
		for _, createdJobID := range createdJobs {
			if pendingJob.ID == createdJobID {
				foundJobs++
				s.a.Equal(job.Pending, pendingJob.Status)
			}
		}
	}
	s.a.Equal(2, foundJobs, "Should find both created jobs in pending status")
}
