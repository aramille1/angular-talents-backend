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

// SendNewEmail sends a verification email using SMTP with Mailtrap templates
func SendNewEmail(templateId, userId, receiverEmail string, code int) error {
	internal.LogInfo("Starting sending sign up confirmation email", map[string]interface{}{"user_id": userId})

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

	// Try to fetch the template from Mailtrap if template ID is provided
	var body string

	if templateId != "" && templateId != "your_mailtrap_template_id" {
		// Define the variables for the template
		templateVars := map[string]string{
			"user_id":           userId,
			"verification_code": fmt.Sprint(code),
			"verification_link": verificationLink,
		}

		// Fetch the rendered template
		renderedTemplate, err := FetchMailtrapTemplate(templateId, templateVars)
		if err == nil {
			// Template was successfully fetched
			internal.LogInfo("Using Mailtrap template for email", map[string]interface{}{
				"template_id": templateId,
			})
			body = renderedTemplate
		} else {
			// If template fetch fails, use the fallback HTML
			internal.LogInfo("Failed to fetch Mailtrap template, using fallback", map[string]interface{}{
				"error": err.Error(),
			})

			// Use fallback HTML template
			body = fmt.Sprintf(`
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
		}
	} else {
		// Use fallback HTML template
		internal.LogInfo("No template ID provided, using fallback email", nil)

		body = fmt.Sprintf(`
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
	}

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

	// Try to fetch the template from Mailtrap if template ID is provided
	var body string
	approvalTemplateID := os.Getenv("RECRUITER_APPROVAL_TEMPLATE_ID")

	if approvalTemplateID != "" && approvalTemplateID != "your_mailtrap_recruiter_approval_template_id" {
		// Define the variables for the template
		templateVars := map[string]string{
			"first_name":   firstName,
			"last_name":    lastName,
			"company":      company,
			"login_url":    loginURL,
			"recruiter_id": recruiterID,
		}

		// Fetch the rendered template
		renderedTemplate, err := FetchMailtrapTemplate(approvalTemplateID, templateVars)
		if err == nil {
			// Template was successfully fetched
			internal.LogInfo("Using Mailtrap template for recruiter approval email", map[string]interface{}{
				"template_id": approvalTemplateID,
			})
			body = renderedTemplate
		} else {
			// If template fetch fails, use the fallback HTML
			internal.LogInfo("Failed to fetch Mailtrap template for recruiter approval, using fallback", map[string]interface{}{
				"error": err.Error(),
			})

			// Use fallback HTML template
			body = fmt.Sprintf(`
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
		}
	} else {
		// Use fallback HTML template
		internal.LogInfo("No recruiter approval template ID provided, using fallback email", nil)

		body = fmt.Sprintf(`
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
	}

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
