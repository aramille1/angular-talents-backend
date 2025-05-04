package domain

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"reverse-job-board/internal"
)

// SendNewEmail sends a verification email using SMTP
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
	from := "hello@angular-talents.mailtrap.io"
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
	from := "hello@angular-talents.mailtrap.io"
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
