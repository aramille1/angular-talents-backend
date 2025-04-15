package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// RecaptchaResponse represents the response from Google's reCAPTCHA verification API
type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

func verifyRecaptcha(recaptchaResponse string) (bool, error) {
	recaptchaSecretKey := os.Getenv("RECAPTCHA_SECRET_KEY")

	// If no secret key is set, skip verification in development
	if recaptchaSecretKey == "" {
		internal.LogInfo("No reCAPTCHA secret key set, skipping verification", nil)
		return true, nil
	}

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{
			"secret":   {recaptchaSecretKey},
			"response": {recaptchaResponse},
		})

	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result RecaptchaResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, err
	}

	return result.Success, nil
}

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

	// Verify reCAPTCHA
	recaptchaValid, err := verifyRecaptcha(userData.RecaptchaResponse)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.validate_recaptcha", "failed to sign up", err.Error())
	}

	if !recaptchaValid {
		return internal.NewError(http.StatusBadRequest, "signup.validate_recaptcha", "failed to sign up", "invalid recaptcha")
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
