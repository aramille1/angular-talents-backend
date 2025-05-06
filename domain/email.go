package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"reverse-job-board/internal"
)

// SendNewEmail sends a verification email using either template ID (API) or SMTP
func SendNewEmail(templateId, userId, receiverEmail string, code int) error {
	internal.LogInfo("Starting sending sign up confirmation email", map[string]interface{}{"user_id": userId})

	// Try to get template ID from environment if not provided
	if templateId == "" {
		templateId = os.Getenv("CONFIRM_EMAIL_TEMPLATE_ID")
	}

	// If we have a valid template ID, use Mailtrap API method
	if templateId != "" && templateId != "your_mailtrap_template_id" {
		return sendEmailWithTemplate(templateId, userId, receiverEmail, code)
	}

	// Otherwise fall back to SMTP method
	return sendEmailWithSMTP(userId, receiverEmail, code)
}

// sendEmailWithTemplate sends an email using Mailtrap API with template
func sendEmailWithTemplate(templateId, userId, receiverEmail string, code int) error {
	internal.LogInfo("Using Mailtrap template for verification email",
		map[string]interface{}{
			"template_id": templateId,
			"user_id":     userId,
		})

	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")
	if mailTrapToken == "" || mailTrapToken == "your_mailtrap_api_token" {
		internal.LogInfo("Mailtrap token not set or using default value", nil)
		return errors.New("mailtrap token not configured properly")
	}

	// Set up frontend URL for the verification link
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}
	verificationLink := fmt.Sprintf("%s/verify/%s/%d", frontendURL, userId, code)

	// Build the request payload
	requestBody := map[string]interface{}{
		"from": map[string]string{
			"email": "hello@angulartalents.com",
			"name":  "Angular Talents",
		},
		"to": []interface{}{
			map[string]string{"email": receiverEmail},
		},
		"template_uuid": templateId,
		"template_variables": map[string]string{
			"user_id":           userId,
			"verification_code": fmt.Sprint(code),
			"verification_link": verificationLink,
		},
	}

	client := &http.Client{}
	url := "https://send.api.mailtrap.io/api/send"

	internal.LogInfo("Preparing Mailtrap API request", map[string]interface{}{
		"api_url":   url,
		"recipient": receiverEmail,
	})

	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		internal.LogInfo("Failed to marshal request body", map[string]interface{}{"error": err.Error()})
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		internal.LogInfo("Failed to create HTTP request", map[string]interface{}{"error": err.Error()})
		return err
	}

	req.Header.Add("Authorization", "Bearer "+mailTrapToken)
	req.Header.Add("Content-Type", "application/json")

	internal.LogInfo("Sending request to Mailtrap API", nil)
	resp, err := client.Do(req)
	if err != nil {
		internal.LogInfo("Failed to send HTTP request", map[string]interface{}{"error": err.Error()})
		return err
	}

	defer resp.Body.Close()
	internal.LogInfo("Received response from Mailtrap API", map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
	})

	if resp.StatusCode != http.StatusOK {
		// Read and log response body for debugging
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		internal.LogInfo("Mailtrap API error response", map[string]interface{}{
			"response_body": bodyString,
		})

		// Create a new buffer with the same content for ReturnRawData
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		err := internal.NewError(http.StatusInternalServerError, "email.send", "request to mailtrap api failed", "failed to send email")
		internal.LogError(err, internal.ReturnRawData(resp))
		return err
	}

	internal.LogInfo("Successfully sent sign up confirmation email using template",
		map[string]interface{}{
			"user_id":     userId,
			"template_id": templateId,
		})
	return nil
}

// sendEmailWithSMTP sends a verification email using SMTP
func sendEmailWithSMTP(userId, receiverEmail string, code int) error {
	internal.LogInfo("Using SMTP for verification email", map[string]interface{}{"user_id": userId})

	// Get SMTP credentials from environment
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "live.smtp.mailtrap.io" // Default Mailtrap SMTP host
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587" // Default Mailtrap SMTP port
	}

	smtpUser := os.Getenv("SMTP_USER")
	if smtpUser == "" {
		smtpUser = "api" // Default Mailtrap SMTP username
	}

	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		smtpPassword = os.Getenv("MAILTRAP_TOKEN") // Try to use existing token as password
	}

	// Check if we have valid configuration
	if smtpPassword == "" || smtpPassword == "your_mailtrap_api_token" {
		internal.LogInfo("SMTP password not set or using default value", nil)
		return errors.New("smtp password not configured properly")
	}

	// Prepare verification email
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}

	// Create the verification link
	verificationLink := fmt.Sprintf("%s/verify/%s/%d", frontendURL, userId, code)

	// Set up email data
	from := "hello@angulartalents.com"
	to := []string{receiverEmail}

	// Create email headers
	subject := "Verify your Angular Talents account"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	// Email body
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify your Angular Talents account</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; border-radius: 5px; padding: 20px; margin-bottom: 20px;">
        <h2 style="color: #333;">Verify your Angular Talents account</h2>
        <p>Thank you for signing up with Angular Talents!</p>
        <p>To verify your email address, please click on the button below:</p>
        <p style="text-align: center;">
            <a href="%s" style="display: inline-block; background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px; margin: 20px 0;">Verify Email</a>
        </p>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all;"><a href="%s">%s</a></p>
        <p>If you did not create an account, you can ignore this email.</p>
        <p>Best regards,<br>The Angular Talents Team</p>
    </div>
</body>
</html>
`, verificationLink, verificationLink, verificationLink)

	// Build the email message
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: Angular Talents <%s>\r\n"+
		"Subject: %s\r\n"+
		"%s\r\n"+
		"%s\r\n",
		receiverEmail,
		from,
		subject,
		mime,
		body))

	// Authentication
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)

	// Send email
	internal.LogInfo("Sending email via SMTP", map[string]interface{}{
		"host": smtpHost,
		"port": smtpPort,
		"user": smtpUser,
	})

	// Connect to the server, authenticate, set the sender and recipient, and send the email
	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		to,
		message,
	)

	if err != nil {
		internal.LogInfo("Failed to send email via SMTP", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	internal.LogInfo("Successfully sent sign up confirmation email via SMTP",
		map[string]interface{}{
			"user_id": userId,
			"email":   receiverEmail,
		})
	return nil
}

// SendRecruiterApprovalEmail sends an email to a recruiter when their profile is approved
func SendRecruiterApprovalEmail(recruiterID, firstName, lastName, company, receiverEmail string) error {
	internal.LogInfo("Starting to send recruiter approval email", map[string]interface{}{"recruiter_id": recruiterID})

	// Try to get approval template ID from environment
	approvalTemplateID := os.Getenv("RECRUITER_APPROVAL_TEMPLATE_ID")

	// If we have a valid template ID, use Mailtrap API method
	if approvalTemplateID != "" && approvalTemplateID != "your_mailtrap_approval_template_id" {
		return sendRecruiterApprovalWithTemplate(approvalTemplateID, recruiterID, firstName, lastName, company, receiverEmail)
	}

	// Otherwise fall back to SMTP method
	return sendRecruiterApprovalWithSMTP(recruiterID, firstName, company, receiverEmail)
}

// sendRecruiterApprovalWithTemplate sends an approval email using Mailtrap API with template
func sendRecruiterApprovalWithTemplate(templateId, recruiterID, firstName, lastName, company, receiverEmail string) error {
	internal.LogInfo("Using Mailtrap template for recruiter approval email",
		map[string]interface{}{
			"template_id":  templateId,
			"recruiter_id": recruiterID,
		})

	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")
	if mailTrapToken == "" || mailTrapToken == "your_mailtrap_api_token" {
		internal.LogInfo("Mailtrap token not set or using default value", nil)
		return errors.New("mailtrap token not configured properly")
	}

	// Set up frontend URL for the login link
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}
	loginURL := fmt.Sprintf("%s/login", frontendURL)

	// Build the request payload
	requestBody := map[string]interface{}{
		"from": map[string]string{
			"email": "hello@angulartalents.com",
			"name":  "Angular Talents",
		},
		"to": []interface{}{
			map[string]string{"email": receiverEmail},
		},
		"template_uuid": templateId,
		"template_variables": map[string]string{
			"first_name":   firstName,
			"last_name":    lastName,
			"company":      company,
			"login_url":    loginURL,
			"recruiter_id": recruiterID,
		},
	}

	client := &http.Client{}
	url := "https://send.api.mailtrap.io/api/send"

	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		internal.LogInfo("Failed to marshal request body", map[string]interface{}{"error": err.Error()})
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		internal.LogInfo("Failed to create HTTP request", map[string]interface{}{"error": err.Error()})
		return err
	}

	req.Header.Add("Authorization", "Bearer "+mailTrapToken)
	req.Header.Add("Content-Type", "application/json")

	internal.LogInfo("Sending request to Mailtrap API", nil)
	resp, err := client.Do(req)
	if err != nil {
		internal.LogInfo("Failed to send HTTP request", map[string]interface{}{"error": err.Error()})
		return err
	}

	defer resp.Body.Close()
	internal.LogInfo("Received response from Mailtrap API", map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
	})

	if resp.StatusCode != http.StatusOK {
		// Read and log response body for debugging
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		internal.LogInfo("Mailtrap API error response", map[string]interface{}{
			"response_body": bodyString,
		})

		// Create a new buffer with the same content for ReturnRawData
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		err := internal.NewError(http.StatusInternalServerError, "email.send_approval", "request to mailtrap api failed", "failed to send approval email")
		internal.LogError(err, internal.ReturnRawData(resp))
		return err
	}

	internal.LogInfo("Successfully sent recruiter approval email using template",
		map[string]interface{}{
			"recruiter_id": recruiterID,
			"template_id":  templateId,
		})
	return nil
}

// sendRecruiterApprovalWithSMTP sends an approval email using SMTP
func sendRecruiterApprovalWithSMTP(recruiterID, firstName, company, receiverEmail string) error {
	internal.LogInfo("Using SMTP for recruiter approval email", map[string]interface{}{"recruiter_id": recruiterID})

	// Get SMTP credentials from environment
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "live.smtp.mailtrap.io" // Default Mailtrap SMTP host
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587" // Default Mailtrap SMTP port
	}

	smtpUser := os.Getenv("SMTP_USER")
	if smtpUser == "" {
		smtpUser = "api" // Default Mailtrap SMTP username
	}

	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		smtpPassword = os.Getenv("MAILTRAP_TOKEN") // Try to use existing token as password
	}

	// Check if we have valid configuration
	if smtpPassword == "" || smtpPassword == "your_mailtrap_api_token" {
		internal.LogInfo("SMTP password not set or using default value", nil)
		return errors.New("smtp password not configured properly")
	}

	// Set up frontend URL
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}

	loginURL := fmt.Sprintf("%s/login", frontendURL)

	// Set up email data
	from := "hello@angulartalents.com"
	to := []string{receiverEmail}

	// Create email headers
	subject := "Your Angular Talents Business Profile has been approved!"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	// Create email body
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Your Business Profile is Approved!</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; border-radius: 5px; padding: 20px; margin-bottom: 20px;">
        <h2 style="color: #333;">Congratulations, %s!</h2>
        <p>Great news! Your business profile for <strong>%s</strong> has been approved.</p>
        <p>You now have full access to the Angular Talents platform and can view all engineers' profiles.</p>
        <p style="text-align: center;">
            <a href="%s" style="display: inline-block; background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px; margin: 20px 0;">Log in now</a>
        </p>
        <p>Start connecting with top Angular talent today!</p>
        <p>Best regards,<br>The Angular Talents Team</p>
    </div>
</body>
</html>
`, firstName, company, loginURL)

	// Build the email message
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: Angular Talents <%s>\r\n"+
		"Subject: %s\r\n"+
		"%s\r\n"+
		"%s\r\n",
		receiverEmail,
		from,
		subject,
		mime,
		body))

	// Authentication
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)

	// Send email
	internal.LogInfo("Sending recruiter approval email via SMTP", map[string]interface{}{
		"host": smtpHost,
		"port": smtpPort,
		"user": smtpUser,
	})

	// Connect to the server, authenticate, set the sender and recipient, and send the email
	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		to,
		message,
	)

	if err != nil {
		internal.LogInfo("Failed to send recruiter approval email via SMTP", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	internal.LogInfo("Successfully sent recruiter approval email via SMTP",
		map[string]interface{}{
			"recruiter_id": recruiterID,
			"email":        receiverEmail,
		})
	return nil
}
