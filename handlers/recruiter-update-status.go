package handlers

import (
	"net/http"
	"reverse-job-board/db"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// UpdateRecruiterStatusRequest represents the request to update a recruiter's status
type UpdateRecruiterStatusRequest struct {
	IsMember bool `json:"is_member"`
}

// HandleRecruiterUpdateStatus handles requests to update a recruiter's membership status
func HandleRecruiterUpdateStatus(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	// Get recruiter ID from URL
	vars := mux.Vars(r.Request)
	recruiterID := vars["recruiterID"]

	// Validate UUID
	_, err := uuid.Parse(recruiterID)
	if err != nil {
		return internal.NewError(
			http.StatusBadRequest,
			"recruiter.update_status.invalid_id",
			"Invalid recruiter ID",
			err.Error(),
		)
	}

	// Parse request body
	var req UpdateRecruiterStatusRequest
	if err := r.DecodeJSON(&w, &req); err != nil {
		return internal.NewError(
			http.StatusBadRequest,
			"recruiter.update_status.invalid_request",
			"Invalid request format",
			err.Error(),
		)
	}

	// Create a custom update payload since `is_member` is not in the standard UpdateRecruiterPayload
	// We'll use a map to update only the is_member field
	updateData := map[string]interface{}{
		"is_member": req.IsMember,
	}

	// Use FindOneAndUpdate directly in handler or modify dao function to accept a map
	recruiterCol := db.Database.Collection("recruiters")
	parsedID, _ := uuid.Parse(recruiterID)

	var updatedRecruiter domain.Recruiter
	err = recruiterCol.FindOneAndUpdate(
		r.Context(),
		map[string]interface{}{"_id": parsedID},
		map[string]interface{}{"$set": updateData},
	).Decode(&updatedRecruiter)

	if err != nil {
		return internal.NewError(
			http.StatusInternalServerError,
			"recruiter.update_status.update_failed",
			"Failed to update recruiter status",
			err.Error(),
		)
	}

	internal.LogInfo("Successfully updated recruiter status", map[string]interface{}{
		"user_id":      r.Context().Value("userID"),
		"recruiter_id": recruiterID,
		"is_member":    req.IsMember,
	})

	w.WriteResponse(http.StatusOK, updatedRecruiter)
	return nil
}
