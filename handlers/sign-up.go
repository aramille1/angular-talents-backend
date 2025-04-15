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
	confirmEmailTemplateId := os.Getenv("CONFIRM_EMAIL_TEMPLATE_ID")
	var userData domain.SignUpData
	err := r.DecodeJSON(&w, &userData)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.decode_body", "failed to sign up", err.Error())
	}

	// Honeypot check - if interests field is filled, it's likely a bot
	if userData.Interests != "" {
		// Return a 200 OK response to make the bot think the signup was successful
		// but don't actually create an account
		internal.LogInfo("Honeypot detected bot attempt", nil)
		w.WriteResponse(http.StatusOK, map[string]string{"status": "success"})
		return nil
	}

	// Verify reCAPTCHA token
	recaptchaResponse, err := domain.VerifyRecaptcha(userData.RecaptchaToken)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "signup.recaptcha_verification", "failed to sign up", err.Error())
	}

	// Check if the score indicates a human user
	if !recaptchaResponse.Success || !domain.IsHumanScore(recaptchaResponse.Score) {
		return internal.NewError(http.StatusBadRequest, "signup.recaptcha_score", "failed to sign up", "reCAPTCHA verification failed")
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
