package handlers

import (
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleManualUserVerify handles requests from admins to manually verify users
func HandleManualUserVerify(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	// Get user ID from URL parameters
	params := mux.Vars(r.Request)
	userID := params["userID"]

	adminID := r.Context().Value("adminID")
	if adminID == nil {
		return internal.NewError(
			http.StatusUnauthorized,
			"user.manual_verify.unauthorized",
			"Unauthorized access",
			"Only admins can manually verify users",
		)
	}

	internal.LogInfo("Starting manual user verification", map[string]interface{}{
		"admin_id":       adminID,
		"target_user_id": userID,
	})

	// Parse user ID
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return internal.NewError(
			http.StatusBadRequest,
			"user.manual_verify.invalid_id",
			"Invalid user ID format",
			err.Error(),
		)
	}

	// Find user in database
	user, err := dao.FindUserById(r.Context(), parsedUserID)
	if err != nil {
		return internal.NewError(
			http.StatusInternalServerError,
			"user.manual_verify.find_user",
			"Failed to find user",
			err.Error(),
		)
	}

	if user == nil {
		return internal.NewError(
			http.StatusNotFound,
			"user.manual_verify.not_found",
			"User not found",
			"No user found with the provided ID",
		)
	}

	// If user is already verified, return success
	if user.Verified {
		internal.LogInfo("User already verified", map[string]interface{}{
			"admin_id":       adminID,
			"target_user_id": userID,
		})
		w.WriteResponse(http.StatusOK, map[string]domain.User{"user": *user})
		return nil
	}

	// Update verification status
	err = dao.UpdateUserVerifiedStatus(r.Context(), userID)
	if err != nil {
		return internal.NewError(
			http.StatusInternalServerError,
			"user.manual_verify.update_failed",
			"Failed to update verification status",
			err.Error(),
		)
	}

	internal.LogInfo("User manually verified", map[string]interface{}{
		"admin_id":       adminID,
		"target_user_id": userID,
	})

	// Get updated user
	updatedUser, err := dao.FindUserById(r.Context(), parsedUserID)
	if err != nil {
		return internal.NewError(
			http.StatusInternalServerError,
			"user.manual_verify.get_updated_user",
			"Failed to retrieve updated user",
			err.Error(),
		)
	}

	w.WriteResponse(http.StatusOK, map[string]domain.User{"user": *updatedUser})
	return nil
}
