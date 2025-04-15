package domain

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reverse-job-board/internal"
	"strings"
)

func SendNewEmail(templateId, userId, receiverEmail string, code int) error {
	internal.LogInfo("Starting sending sign up confirmation email", map[string]interface{}{"user_id": userId})

	// Get Mailgun credentials from environment variables
	mailgunDomain := os.Getenv("MAILGUN_DOMAIN")
	mailgunAPIKey := os.Getenv("MAILGUN_API_KEY")

	if mailgunDomain == "" || mailgunAPIKey == "" {
		// For development: just log the verification link and return success
		internal.LogInfo("Mailgun credentials not found, skipping actual email send", map[string]interface{}{
			"user_id":           userId,
			"verification_link": fmt.Sprintf("http://localhost:8080/verify/%s/%d", userId, code),
			"email":             receiverEmail,
		})
		return nil
	}

	// The verification URL that will be linked in the email
	verificationURL := fmt.Sprintf("https://angulartalents.onrender.com/verify/%s/%d", userId, code)

	// Build form data for the Mailgun API
	formData := url.Values{}
	formData.Add("from", "The Angular Team <hello@angulartalents.com>")
	formData.Add("to", receiverEmail)
	formData.Add("subject", "Verify Your Angular Talents Account")

	emailHTML := fmt.Sprintf(`
		<h2>Welcome to Angular Talents!</h2>
		<p>Thank you for signing up. Please verify your email address by clicking the link below:</p>
		<p><a href="%s">Verify Your Email</a></p>
		<p>Or copy and paste this URL into your browser:</p>
		<p>%s</p>
		<p>If you did not create an account, you can safely ignore this email.</p>
		<p>Best regards,<br>The Angular Team</p>
	`, verificationURL, verificationURL)

	formData.Add("html", emailHTML)

	// Create the request
	client := &http.Client{}
	mailgunURL := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", mailgunDomain)

	req, err := http.NewRequest("POST", mailgunURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth("api", mailgunAPIKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := internal.NewError(http.StatusInternalServerError, "email.send", "request to Mailgun API failed", fmt.Sprintf("failed to send email, status code: %d", resp.StatusCode))
		internal.LogError(err, internal.ReturnRawData(resp))
		return err
	}

	internal.LogInfo("Successfully sent sign up confirmation email", map[string]interface{}{"user_id": userId})
	return nil
}
