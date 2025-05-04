package handlers

import (
	"net/http"
	"os"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/go-playground/validator/v10"
)

func HandleSignUp(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	internal.LogInfo("Starting sign up", nil)
	confirmEmailTemplateId := os.Getenv("CONFIRM_EMAIL_TEMPLATE_ID")
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

	// Check honeypot field - if filled, it's likely a bot
	if userData.Website != "" {
		internal.LogInfo("Honeypot trap triggered", map[string]interface{}{"honeypot_value": userData.Website})
		return internal.NewError(http.StatusBadRequest, "signup.honeypot", "failed to sign up", "bot detected")
	}

	user, err := userData.NewUser()
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.create_user", "failed to sign up", err.Error())
	}

	err = user.Validate(r.Context())
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "signup.validate_user", "failed to sign up", err.Error())
	}

	userId, err := dao.InsertNewUser(r.Context(), user)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.insert_user", "failed to sign up", err.Error())
	}

	// Send Slack notification (non-blocking)
	go internal.NotifyNewUserRegistration(user.Email, user.ID.String())

	// Try to send the verification email, but don't fail the registration if it fails
	err = domain.SendNewEmail(confirmEmailTemplateId, userId, user.Email, user.VerificationCode)
	if err != nil {
		internal.LogError(
			internal.NewError(http.StatusInternalServerError, "signup.send_confirmation_email", "request to mailtrap api failed", err.Error()),
			map[string]interface{}{"user_id": user.ID, "email": user.Email},
		)

		// Still return a specific error about the email sending failure, but with a 201 Created status
		// to indicate the account was created but we couldn't send the verification email
		w.WriteResponse(http.StatusCreated, map[string]domain.User{"user": *user})

		return internal.NewError(http.StatusInternalServerError, "signup.send_confirmation_email", "failed to sign up", "request to mailtrap api failed")
	}

	internal.LogInfo("Successfully signed up user", map[string]interface{}{"user_id": user.ID})
	w.WriteResponse(http.StatusOK, map[string]domain.User{"user": *user})
	return nil
}
