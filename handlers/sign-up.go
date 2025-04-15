package handlers

import (
	"net/http"
	"os"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func HandleSignUp(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	internal.LogInfo("Starting sign up", nil)
	var userData domain.SignUpData
	err := r.DecodeJSON(&w, &userData)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.decode_body", "failed to sign up", err.Error())
	}

	v := validator.New()
	err = v.Struct(userData)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "signup.validate_body", "failed to sign up", err.Error())
	}

	user, err := userData.NewUser()
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.create_user", "failed to sign up", err.Error())
	}

	err = user.Validate(r.Context())
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "signup.validate_user", "failed to sign up", err.Error())
	}

	// For development: auto-verify users if no Mailgun credentials are present
	if os.Getenv("MAILGUN_API_KEY") == "" || os.Getenv("MAILGUN_DOMAIN") == "" {
		user.Verified = true
	}

	userId, err := dao.InsertNewUser(r.Context(), user)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.insert_user", "failed to sign up", err.Error())
	}

	// Send verification email
	err = domain.SendNewEmail("", userId, user.Email, user.VerificationCode)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.send_confirmation_email", "failed to sign up", err.Error())
	}

	internal.LogInfo("Successfully signed up user", map[string]interface{}{"user_id": user.ID})
	w.WriteResponse(http.StatusOK, map[string]uuid.UUID{"user_id": user.ID})
	return nil
}
