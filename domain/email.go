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

// FetchMailtrapTemplate retrieves the rendered template from Mailtrap API
func FetchMailtrapTemplate(templateId string, variables map[string]string) (string, error) {
	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")

	// Check if we have valid configuration
	if mailTrapToken == "" || mailTrapToken == "your_mailtrap_api_token" {
		internal.LogInfo("Mailtrap token not set or using default value", nil)
		return "", errors.New("mailtrap token not configured properly")
	}

	// Build the request body for template preview
	requestBody := map[string]interface{}{
		"template_uuid":      templateId,
		"template_variables": variables,
	}

	client := &http.Client{}
	url := "https://api.mailtrap.io/api/render_template"

	internal.LogInfo("Fetching Mailtrap template", map[string]interface{}{
		"template_id": templateId,
		"variables":   variables,
	})

	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		internal.LogInfo("Failed to marshal template request body", map[string]interface{}{"error": err.Error()})
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		internal.LogInfo("Failed to create template HTTP request", map[string]interface{}{"error": err.Error()})
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+mailTrapToken)
	req.Header.Add("Content-Type", "application/json")

	internal.LogInfo("Sending request to Mailtrap API for template", nil)
	resp, err := client.Do(req)
	if err != nil {
		internal.LogInfo("Failed to send template HTTP request", map[string]interface{}{"error": err.Error()})
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		internal.LogInfo("Mailtrap template API error response", map[string]interface{}{
			"status_code":   resp.StatusCode,
			"response_body": bodyString,
		})
		return "", fmt.Errorf("failed to fetch template: %s", bodyString)
	}

	// Parse the response
	var templateResp struct {
		HTML string `json:"html"`
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		internal.LogInfo("Failed to read template response body", map[string]interface{}{"error": err.Error()})
		return "", err
	}

	err = json.Unmarshal(bodyBytes, &templateResp)
	if err != nil {
		internal.LogInfo("Failed to unmarshal template response", map[string]interface{}{"error": err.Error()})
		return "", err
	}

	return templateResp.HTML, nil
}

// SendNewEmail sends a verification email using SMTP with Mailtrap template
func SendNewEmail(templateId, userId, receiverEmail string, code int) error {
	internal.LogInfo("Starting sending sign up confirmation email", map[string]interface{}{"user_id": userId})

	// Check if template ID is configured
	if templateId == "" || templateId == "your_mailtrap_template_id" {
		internal.LogInfo("Mailtrap template ID not set, using fallback email", nil)
		return sendFallbackVerificationEmail(receiverEmail, userId, code)
	}

	// First fetch the rendered template from Mailtrap API
	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")
	if mailTrapToken == "" || mailTrapToken == "your_mailtrap_api_token" {
		internal.LogInfo("Mailtrap token not set, using fallback email", nil)
		return sendFallbackVerificationEmail(receiverEmail, userId, code)
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}

	// Create the verification link
	verificationLink := fmt.Sprintf("%s/verify/%s/%d", frontendURL, userId, code)

	// Prepare API request to render template
	requestBody := map[string]interface{}{
		"template_uuid": templateId,
		"template_variables": map[string]string{
			"user_id":           userId,
			"verification_code": fmt.Sprint(code),
			"verification_link": verificationLink,
		},
	}

	internal.LogInfo("Fetching template from Mailtrap API", map[string]interface{}{
		"template_id": templateId,
	})

	// Convert request body to JSON
	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		internal.LogInfo("Failed to marshal template request", map[string]interface{}{"error": err.Error()})
		return sendFallbackVerificationEmail(receiverEmail, userId, code)
	}

	// Create HTTP request to fetch the template
	client := &http.Client{}
	renderURL := "https://templates.api.mailtrap.io/api/render"
	req, err := http.NewRequest("POST", renderURL, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		internal.LogInfo("Failed to create template fetch request", map[string]interface{}{"error": err.Error()})
		return sendFallbackVerificationEmail(receiverEmail, userId, code)
	}

	// Set authorization headers
	req.Header.Add("Authorization", "Bearer "+mailTrapToken)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		internal.LogInfo("Failed to fetch template", map[string]interface{}{"error": err.Error()})
		return sendFallbackVerificationEmail(receiverEmail, userId, code)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		internal.LogInfo("Mailtrap template API error", map[string]interface{}{
			"status_code": resp.StatusCode,
		})
		return sendFallbackVerificationEmail(receiverEmail, userId, code)
	}

	// Read the template HTML from the response
	var templateResponse struct {
		HTML string `json:"html"`
		Text string `json:"text"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		internal.LogInfo("Failed to read template response", map[string]interface{}{"error": err.Error()})
		return sendFallbackVerificationEmail(receiverEmail, userId, code)
	}

	err = json.Unmarshal(body, &templateResponse)
	if err != nil {
		internal.LogInfo("Failed to parse template response", map[string]interface{}{"error": err.Error()})
		return sendFallbackVerificationEmail(receiverEmail, userId, code)
	}

	// Now we have the template HTML, send via SMTP
	return sendEmailViaSmtp(receiverEmail, "Verify your Angular Talents account", templateResponse.HTML)
}

// SendRecruiterApprovalEmail sends an email to a recruiter when their profile is approved
func SendRecruiterApprovalEmail(recruiterID, firstName, lastName, company, receiverEmail string) error {
	internal.LogInfo("Starting to send recruiter approval email", map[string]interface{}{"recruiter_id": recruiterID})

	approvalTemplateID := os.Getenv("RECRUITER_APPROVAL_TEMPLATE_ID")
	if approvalTemplateID == "" {
		internal.LogInfo("Recruiter approval template ID not set, using fallback method", nil)
		return sendFallbackApprovalEmail(receiverEmail, firstName, company)
	}

	// First fetch the rendered template from Mailtrap API
	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")
	if mailTrapToken == "" || mailTrapToken == "your_mailtrap_api_token" {
		internal.LogInfo("Mailtrap token not set, using fallback email", nil)
		return sendFallbackApprovalEmail(receiverEmail, firstName, company)
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}

	loginURL := fmt.Sprintf("%s/login", frontendURL)

	// Prepare API request to render template
	requestBody := map[string]interface{}{
		"template_uuid": approvalTemplateID,
		"template_variables": map[string]string{
			"first_name":   firstName,
			"last_name":    lastName,
			"company":      company,
			"login_url":    loginURL,
			"recruiter_id": recruiterID,
		},
	}

	internal.LogInfo("Fetching approval template from Mailtrap API", map[string]interface{}{
		"template_id": approvalTemplateID,
	})

	// Convert request body to JSON
	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		internal.LogInfo("Failed to marshal approval template request", map[string]interface{}{"error": err.Error()})
		return sendFallbackApprovalEmail(receiverEmail, firstName, company)
	}

	// Create HTTP request to fetch the template
	client := &http.Client{}
	renderURL := "https://templates.api.mailtrap.io/api/render"
	req, err := http.NewRequest("POST", renderURL, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		internal.LogInfo("Failed to create approval template fetch request", map[string]interface{}{"error": err.Error()})
		return sendFallbackApprovalEmail(receiverEmail, firstName, company)
	}

	// Set authorization headers
	req.Header.Add("Authorization", "Bearer "+mailTrapToken)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		internal.LogInfo("Failed to fetch approval template", map[string]interface{}{"error": err.Error()})
		return sendFallbackApprovalEmail(receiverEmail, firstName, company)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		internal.LogInfo("Mailtrap approval template API error", map[string]interface{}{
			"status_code": resp.StatusCode,
		})
		return sendFallbackApprovalEmail(receiverEmail, firstName, company)
	}

	// Read the template HTML from the response
	var templateResponse struct {
		HTML string `json:"html"`
		Text string `json:"text"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		internal.LogInfo("Failed to read approval template response", map[string]interface{}{"error": err.Error()})
		return sendFallbackApprovalEmail(receiverEmail, firstName, company)
	}

	err = json.Unmarshal(body, &templateResponse)
	if err != nil {
		internal.LogInfo("Failed to parse approval template response", map[string]interface{}{"error": err.Error()})
		return sendFallbackApprovalEmail(receiverEmail, firstName, company)
	}

	// Now we have the template HTML, send via SMTP
	return sendEmailViaSmtp(receiverEmail, "Your Angular Talents Business Profile has been approved!", templateResponse.HTML)
}

// sendEmailViaSmtp sends an email using SMTP
func sendEmailViaSmtp(receiverEmail, subject, htmlBody string) error {
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

	// Set up email data
	from := "hello@angulartalents.com"
	to := []string{receiverEmail}

	// Create email headers
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

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
		htmlBody))

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

	internal.LogInfo("Successfully sent email via SMTP", map[string]interface{}{
		"to": receiverEmail,
	})
	return nil
}

// sendFallbackVerificationEmail sends a basic verification email when template rendering fails
func sendFallbackVerificationEmail(receiverEmail, userId string, code int) error {
	internal.LogInfo("Using fallback verification email method", map[string]interface{}{
		"email":   receiverEmail,
		"user_id": userId,
	})

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}

	internal.LogInfo("Fallback email configuration", map[string]interface{}{
		"frontend_url": frontendURL,
	})

	verificationLink := fmt.Sprintf("%s/verify/%s/%d", frontendURL, userId, code)

	// HTML version of the email
	html := fmt.Sprintf(`
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

	// Send the fallback email via SMTP
	return sendEmailViaSmtp(receiverEmail, "Verify your Angular Talents account", html)
}

// sendFallbackApprovalEmail sends a basic approval email when template rendering fails
func sendFallbackApprovalEmail(receiverEmail, firstName, company string) error {
	internal.LogInfo("Using fallback approval email method", map[string]interface{}{
		"email":        receiverEmail,
		"first_name":   firstName,
		"company_name": company,
	})

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}

	loginURL := fmt.Sprintf("%s/login", frontendURL)

	// HTML version of the email
	html := fmt.Sprintf(`
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

	// Send the fallback email via SMTP
	return sendEmailViaSmtp(receiverEmail, "Your Angular Talents Business Profile has been approved!", html)
}
