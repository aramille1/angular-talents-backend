package handlers

import (
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/internal"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleGetUserEmail handles requests to get a user's email by their ID
func HandleGetUserEmail(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	// Get user ID from URL parameters
	params := mux.Vars(r.Request)
	userID := params["userID"]

	internal.LogInfo("Starting get user email", map[string]interface{}{
		"admin_id":       r.Context().Value("adminID"),
		"target_user_id": userID,
	})

	// Parse user ID
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return internal.NewError(
			http.StatusBadRequest,
			"user.email.invalid_id",
			"Invalid user ID format",
			err.Error(),
		)
	}

	// Find user in database
	user, err := dao.FindUserById(r.Context(), parsedUserID)
	if err != nil {
		return internal.NewError(
			http.StatusInternalServerError,
			"user.email.find_user",
			"Failed to find user",
			err.Error(),
		)
	}

	if user == nil {
		return internal.NewError(
			http.StatusNotFound,
			"user.email.not_found",
			"User not found",
			"No user found with the provided ID",
		)
	}

	// Return email and created_at
	response := map[string]interface{}{
		"email": user.Email,
	}

	// Only include createdAt if it's set
	if !user.CreatedAt.IsZero() {
		response["createdAt"] = user.CreatedAt
	}

	w.WriteResponse(http.StatusOK, response)
	return nil
}
