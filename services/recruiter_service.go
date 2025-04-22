package services

import (
	"context"
	"errors"
	"time"

	"github.com/go-mongodb/mongodb"
	"github.com/go-mongodb/mongodb/bson"
	"github.com/go-mongodb/mongodb/primitive"

	"github.com/angular-talents-backend/models"
)

type RecruiterService struct {
	recruiterRepo *mongodb.Collection
}

// GetRecruiters retrieves recruiters with filtering by status and pagination
func (s *RecruiterService) GetRecruiters(ctx context.Context, page, limit int, status string) ([]*models.RecruiterResponse, int64, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	// Calculate skip (for pagination)
	skip := (page - 1) * limit

	// Get total count
	total, err := s.recruiterRepo.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.New("error counting recruiters")
	}

	// Get recruiters
	recruiters, err := s.recruiterRepo.FindWithFilter(ctx, filter, skip, limit)
	if err != nil {
		return nil, 0, errors.New("error retrieving recruiters")
	}

	// Convert to response objects
	var recruiterResponses []*models.RecruiterResponse
	for _, recruiter := range recruiters {
		resp := models.NewRecruiterResponse(recruiter)
		recruiterResponses = append(recruiterResponses, &resp)
	}

	return recruiterResponses, total, nil
}

// GetRecruitersByStatus retrieves recruiters by status
func (s *RecruiterService) GetRecruitersByStatus(ctx context.Context, status string) ([]*models.RecruiterResponse, error) {
	// Get recruiters
	recruiters, err := s.recruiterRepo.FindByStatus(ctx, status)
	if err != nil {
		return nil, errors.New("error retrieving recruiters")
	}

	// Convert to response objects
	var recruiterResponses []*models.RecruiterResponse
	for _, recruiter := range recruiters {
		resp := models.NewRecruiterResponse(recruiter)
		recruiterResponses = append(recruiterResponses, &resp)
	}

	return recruiterResponses, nil
}

// UpdateRecruiterStatus updates a recruiter's status (approve/reject)
func (s *RecruiterService) UpdateRecruiterStatus(
	ctx context.Context,
	id string,
	status string,
	adminID string,
	rejectionReason string,
) (*models.RecruiterResponse, error) {
	// Parse recruiter ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid recruiter ID format")
	}

	// Get existing recruiter
	recruiter, err := s.recruiterRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, errors.New("recruiter not found")
	}

	// Parse admin ID
	adminObjID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return nil, errors.New("invalid admin ID format")
	}

	// Update status
	recruiter.Status = status

	// Update additional fields
	if status == "approved" {
		recruiter.AdminVerified = true
		recruiter.RejectionReason = ""
	} else if status == "rejected" {
		recruiter.AdminVerified = false
		recruiter.RejectionReason = rejectionReason
	}

	recruiter.ApprovedBy = adminObjID
	recruiter.ApprovalDate = time.Now()

	// Save updates
	if err := s.recruiterRepo.Update(ctx, recruiter); err != nil {
		return nil, errors.New("error updating recruiter")
	}

	// Convert to response
	resp := models.NewRecruiterResponse(recruiter)
	return &resp, nil
}
