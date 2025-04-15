package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reverse-job-board/internal"
)

// RecaptchaResponse represents the response from Google's reCAPTCHA API
type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

// VerifyRecaptcha verifies a reCAPTCHA token with Google's API
func VerifyRecaptcha(token string) (*RecaptchaResponse, error) {
	recaptchaSecret := os.Getenv("RECAPTCHA_SECRET_KEY")
	if recaptchaSecret == "" {
		return nil, errors.New("RECAPTCHA_SECRET_KEY not set in environment")
	}

	requestBody := map[string]string{
		"secret":   recaptchaSecret,
		"response": token,
	}

	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://www.google.com/recaptcha/api/siteverify", bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := internal.NewError(http.StatusInternalServerError, "recaptcha.verify", "request to recaptcha api failed", "failed to verify recaptcha")
		internal.LogError(err, internal.ReturnRawData(resp))
		return nil, err
	}

	var recaptchaResponse RecaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&recaptchaResponse); err != nil {
		return nil, err
	}

	return &recaptchaResponse, nil
}

// IsHumanScore checks if the reCAPTCHA score is high enough to be considered human
func IsHumanScore(score float64) bool {
	// Threshold can be adjusted based on your requirements
	// 0.5 is a moderate threshold, 0.7 is more strict, 0.3 is more lenient
	threshold := 0.5
	return score >= threshold
}
