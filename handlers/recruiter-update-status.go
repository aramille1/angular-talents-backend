package handlers

import (
	"net/http"
	"reverse-job-board/db"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
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
	parsedID, err := uuid.Parse(recruiterID)
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

	// Fetch the recruiter to get their details
	recruiterCol := db.Database.Collection("recruiters")

	// Get the recruiter before updating it
	var recruiter domain.Recruiter
	err = recruiterCol.FindOne(
		r.Context(),
		bson.M{"_id": parsedID},
	).Decode(&recruiter)

	if err != nil {
		return internal.NewError(
			http.StatusInternalServerError,
			"recruiter.update_status.find_failed",
			"Failed to find recruiter",
			err.Error(),
		)
	}

	// Store the previous status to detect status change
	wasApproved := recruiter.IsMember

	// Create a custom update payload since `is_member` is not in the standard UpdateRecruiterPayload
	// We'll use a map to update only the is_member field
	updateData := map[string]interface{}{
		"is_member": req.IsMember,
	}

	// Update the recruiter status
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

	// If the status changed from not approved to approved, send an email
	if !wasApproved && req.IsMember {
		// Get the associated user to retrieve the email
		usersCol := db.Database.Collection("users")
		var user domain.User

		err = usersCol.FindOne(
			r.Context(),
			bson.M{"_id": recruiter.UserID},
		).Decode(&user)

		if err != nil {
			internal.LogInfo("Failed to find user for approved recruiter notification", map[string]interface{}{
				"recruiter_id": recruiterID,
				"user_id":      recruiter.UserID.String(),
				"error":        err.Error(),
			})
		} else {
			// Send email notification asynchronously
			go func() {
				err := domain.SendRecruiterApprovalEmail(
					recruiterID,
					recruiter.Firstname,
					recruiter.Lastname,
					recruiter.Company,
					user.Email,
				)

				if err != nil {
					internal.LogInfo("Failed to send approval email", map[string]interface{}{
						"recruiter_id": recruiterID,
						"error":        err.Error(),
					})
				}
			}()
		}
	}

	internal.LogInfo("Successfully updated recruiter status", map[string]interface{}{
		"user_id":      r.Context().Value("userID"),
		"recruiter_id": recruiterID,
		"is_member":    req.IsMember,
	})

	w.WriteResponse(http.StatusOK, updatedRecruiter)
	return nil
}
