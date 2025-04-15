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

// SignUpRequest combines user data with honeypot field
type SignUpRequest struct {
	domain.SignUpData
	// Email2 is a honeypot field that should remain empty
	Email2 string `json:"email2"`
}

func HandleSignUp(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	internal.LogInfo("Starting sign up", nil)
	confirmEmailTemplateId := os.Getenv("CONFIRM_EMAIL_TEMPLATE_ID")

	// Decode the full request with both user data and honeypot
	var signupRequest SignUpRequest
	err := r.DecodeJSON(&w, &signupRequest)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.decode_body", "failed to sign up", err.Error())
	}

	// Check honeypot field - if it's filled, it's likely a bot
	if signupRequest.Email2 != "" {
		// Return success response to not alert the bot, but don't proceed with registration
		internal.LogInfo("Honeypot triggered, likely bot", nil)
		w.WriteResponse(http.StatusOK, map[string]string{"message": "Check your email for verification instructions"})
		return nil
	}

	// Extract the user data part
	userData := signupRequest.SignUpData

	// Validate form data
	v := validator.New()
	err = v.Struct(userData)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "signup.validate_body", "failed to sign up", err.Error())
	}

	// Verify reCAPTCHA token
	recaptchaValid, err := internal.VerifyRecaptcha(userData.RecaptchaToken)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "signup.verify_recaptcha", "failed to sign up", err.Error())
	}

	if !recaptchaValid {
		return internal.NewError(http.StatusBadRequest, "signup.invalid_recaptcha", "failed to sign up", "Invalid reCAPTCHA")
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

	err = domain.SendNewEmail(confirmEmailTemplateId, userId, user.Email, user.VerificationCode)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.send_confirmation_email", "failed to sign up", err.Error())
	}

	internal.LogInfo("Successfully signed up user", map[string]interface{}{"user_id": user.ID})
	w.WriteResponse(http.StatusOK, map[string]uuid.UUID{"user_id": user.ID})
	return nil
}
