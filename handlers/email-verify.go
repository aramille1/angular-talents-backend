package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reverse-job-board/dao"
	"reverse-job-board/internal"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func HandleEmailVerify(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	params := mux.Vars(r.Request)
	userId := params["userID"]
	verificationCode := params["verificationCode"]
	internal.LogInfo("Starting email verification", map[string]interface{}{"user_id": userId, "code": verificationCode})

	// Parse user ID
	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		internal.LogError(internal.NewError(http.StatusBadRequest, "email.verify.parse_user_id", "failed to verify email", err.Error()), nil)
		redirectToErrorPage(w, "Invalid verification link")
		return nil
	}

	// Find user by ID
	user, err := dao.FindUserById(r.Context(), parsedUserId)
	if err != nil {
		internal.LogError(internal.NewError(http.StatusInternalServerError, "email.verify.read_user_by_id", "failed to verify email", err.Error()), nil)
		redirectToErrorPage(w, "Failed to find user")
		return nil
	}

	if user == nil {
		internal.LogError(internal.NewError(http.StatusNotFound, "email.verify.user_not_found", "failed to verify email", "user not found"), nil)
		redirectToErrorPage(w, "User not found")
		return nil
	}

	// Check if user is already verified
	if user.Verified {
		// User is already verified, redirect to success page
		internal.LogInfo("User already verified, redirecting to success page", map[string]interface{}{"user_id": userId})
		redirectToSuccessPage(w)
		return nil
	}

	// Verify the code
	if fmt.Sprint(user.VerificationCode) != verificationCode {
		internal.LogError(internal.NewError(http.StatusBadRequest, "email.verify.check_verification_code", "failed to verify email", "verification code incorrect"), nil)
		redirectToErrorPage(w, "Invalid verification code")
		return nil
	}

	// Update user verification status
	err = dao.UpdateUserVerifiedStatus(r.Context(), userId)
	if err != nil {
		internal.LogError(internal.NewError(http.StatusInternalServerError, "email.verify.update_user_verified_status", "failed to verify email", err.Error()), nil)
		redirectToErrorPage(w, "Failed to verify email")
		return nil
	}

	internal.LogInfo("Successfully verified email", map[string]interface{}{"user_id": userId})

	// Redirect to success page
	redirectToSuccessPage(w)
	return nil
}

// redirectToSuccessPage redirects to the frontend success page
func redirectToSuccessPage(w http.ResponseWriter) {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://angulartalents.onrender.com" // Default frontend URL
	}

	// Redirect to the verification success page
	successURL := fmt.Sprintf("%s/verification-success", frontendURL)
	w.Header().Set("Location", successURL)
	w.WriteHeader(http.StatusSeeOther)
}

// redirectToErrorPage redirects to the frontend error page with a message
func redirectToErrorPage(w http.ResponseWriter, message string) {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://angulartalents.onrender.com" // Default frontend URL
	}

	// Redirect to the verification error page with the error message
	errorURL := fmt.Sprintf("%s/verification-error?message=%s", frontendURL, url.QueryEscape(message))
	w.Header().Set("Location", errorURL)
	w.WriteHeader(http.StatusSeeOther)
}
