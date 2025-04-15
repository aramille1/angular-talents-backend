package internal

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// RecaptchaResponse represents the response from Google's reCAPTCHA verification API
type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

// VerifyRecaptcha validates a reCAPTCHA token against Google's verification API
func VerifyRecaptcha(token string) (bool, error) {
	// Skip verification in development mode if enabled
	if os.Getenv("SKIP_RECAPTCHA") == "true" {
		return true, nil
	}

	if token == "" {
		return false, errors.New("missing recaptcha token")
	}

	recaptchaSecret := os.Getenv("RECAPTCHA_SECRET_KEY")
	if recaptchaSecret == "" {
		LogError(NewError(http.StatusInternalServerError, "recaptcha.verify", "recaptcha configuration error", "RECAPTCHA_SECRET_KEY not set"), nil)
		return false, errors.New("recaptcha configuration error")
	}

	// Prepare form data
	formData := url.Values{
		"secret":   {recaptchaSecret},
		"response": {token},
	}

	// Make the HTTP request to Google's reCAPTCHA API
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", formData)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Decode response
	var recaptchaResp RecaptchaResponse
	err = json.NewDecoder(resp.Body).Decode(&recaptchaResp)
	if err != nil {
		return false, err
	}

	// Check if verification was successful
	if !recaptchaResp.Success {
		errorMsg := strings.Join(recaptchaResp.ErrorCodes, ", ")
		return false, errors.New("recaptcha verification failed: " + errorMsg)
	}

	return true, nil
}

// CheckHoneypot checks if a honeypot field was filled (which bots would do)
func CheckHoneypot(value string) bool {
	// If the honeypot field is empty, this is likely a human
	// If it's filled, it's likely a bot
	return value == ""
}
