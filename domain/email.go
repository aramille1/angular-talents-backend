package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reverse-job-board/internal"
)

func SendNewEmail(templateId, userId, receiverEmail string, code int) error {
	internal.LogInfo("Starting sending sign up confirmation email", map[string]interface{}{"user_id": userId})
	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")

	// Debug logging for token
	tokenDebug := "not set"
	if mailTrapToken != "" {
		if len(mailTrapToken) > 8 {
			// Show first 4 and last 4 chars only
			tokenDebug = mailTrapToken[:4] + "..." + mailTrapToken[len(mailTrapToken)-4:]
		} else {
			tokenDebug = mailTrapToken[:2] + "..."
		}
	}

	internal.LogInfo("Mailtrap configuration", map[string]interface{}{
		"token_sample": tokenDebug,
		"template_id":  templateId,
		"token_length": len(mailTrapToken),
	})

	// Check if we have valid configuration
	if mailTrapToken == "" || mailTrapToken == "your_mailtrap_api_token" {
		internal.LogInfo("Mailtrap token not set or using default value", nil)
		return errors.New("mailtrap token not configured properly")
	}

	if templateId == "" || templateId == "your_mailtrap_template_id" {
		internal.LogInfo("Mailtrap template ID not set or using default value, using fallback email", nil)
		return sendFallbackVerificationEmail(receiverEmail, userId, code)
	}

	requestBody := map[string]interface{}{
		"from": map[string]string{
			"email": "hello.angulartalents@gmail.com",
			"name":  "The Angular Team",
		},
		"to": []interface{}{
			map[string]string{"email": receiverEmail},
		},
		"template_uuid": templateId,
		"template_variables": map[string]string{
			"user_id":           userId,
			"verification_code": fmt.Sprint(code),
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

	// Debug log the authorization header (partially)
	authHeader := "Bearer " + mailTrapToken
	authDebug := "Bearer " + tokenDebug
	internal.LogInfo("Setting authorization header", map[string]interface{}{"auth_header_sample": authDebug})

	req.Header.Add("Authorization", authHeader)
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

	internal.LogInfo("Successfully sent sign up confirmation email", map[string]interface{}{"user_id": userId})
	return nil
}

// SendRecruiterApprovalEmail sends an email to a recruiter when their profile is approved
func SendRecruiterApprovalEmail(recruiterID, firstName, lastName, company, receiverEmail string) error {
	internal.LogInfo("Starting to send recruiter approval email", map[string]interface{}{"recruiter_id": recruiterID})

	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")
	approvalTemplateID := os.Getenv("RECRUITER_APPROVAL_TEMPLATE_ID")

	if approvalTemplateID == "" {
		internal.LogInfo("Recruiter approval template ID not set, using fallback method", nil)
		return sendFallbackApprovalEmail(receiverEmail, firstName, company)
	}

	requestBody := map[string]interface{}{
		"from": map[string]string{
			"email": "hello.angulartalents@gmail.com",
			"name":  "Angular Talents",
		},
		"to": []interface{}{
			map[string]string{"email": receiverEmail},
		},
		"template_uuid": approvalTemplateID,
		"template_variables": map[string]string{
			"first_name":   firstName,
			"last_name":    lastName,
			"company":      company,
			"login_url":    fmt.Sprintf("%s/login", os.Getenv("FRONTEND_URL")),
			"recruiter_id": recruiterID,
		},
	}

	client := &http.Client{}
	url := "https://send.api.mailtrap.io/api/send"

	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+mailTrapToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := internal.NewError(http.StatusInternalServerError, "email.send_approval", "request to mailtrap api failed", "failed to send approval email")
		internal.LogError(err, internal.ReturnRawData(resp))
		return err
	}

	internal.LogInfo("Successfully sent recruiter approval email", map[string]interface{}{"recruiter_id": recruiterID})
	return nil
}

// sendFallbackApprovalEmail sends a basic approval email when template ID is not configured
func sendFallbackApprovalEmail(receiverEmail, firstName, company string) error {
	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")

	// Simple subject and text for fallback email
	subject := "Your Angular Talents Business Profile has been approved!"
	text := fmt.Sprintf("Hello %s,\n\nGreat news! Your business profile for %s has been approved. You now have full access to the Angular Talents platform and can view all engineers' profiles.\n\nLog in now to start connecting with top Angular talent: %s/login\n\nBest regards,\nThe Angular Talents Team",
		firstName,
		company,
		os.Getenv("FRONTEND_URL"),
	)

	// Build the request body for a simple email without a template
	requestBody := map[string]interface{}{
		"from": map[string]string{
			"email": "hello.angulartalents@gmail.com",
			"name":  "Angular Talents",
		},
		"to": []interface{}{
			map[string]string{"email": receiverEmail},
		},
		"subject": subject,
		"text":    text,
	}

	client := &http.Client{}
	url := "https://send.api.mailtrap.io/api/send"

	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+mailTrapToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := internal.NewError(http.StatusInternalServerError, "email.send_fallback_approval", "request to mailtrap api failed", "failed to send fallback approval email")
		internal.LogError(err, internal.ReturnRawData(resp))
		return err
	}

	internal.LogInfo("Successfully sent fallback recruiter approval email", map[string]interface{}{"email": receiverEmail})
	return nil
}

// sendFallbackVerificationEmail sends a basic verification email when template ID is not configured
func sendFallbackVerificationEmail(receiverEmail, userId string, code int) error {
	internal.LogInfo("Using fallback verification email method", map[string]interface{}{
		"email":   receiverEmail,
		"user_id": userId,
	})

	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}

	internal.LogInfo("Fallback email configuration", map[string]interface{}{
		"frontend_url": frontendURL,
	})

	verificationLink := fmt.Sprintf("%s/verify/%s/%d", frontendURL, userId, code)

	// Simple subject and text for fallback email
	subject := "Verify your Angular Talents account"
	text := fmt.Sprintf("Hello,\n\nThank you for signing up with Angular Talents! To verify your email address, please click on the link below:\n\n%s\n\nIf you did not create an account, you can ignore this email.\n\nBest regards,\nThe Angular Talents Team",
		verificationLink,
	)

	// HTML version of the email
	html := fmt.Sprintf(`
	<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
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
	`, verificationLink, verificationLink, verificationLink)

	// Build the request body for a simple email without a template
	requestBody := map[string]interface{}{
		"from": map[string]string{
			"email": "hello.angulartalents@gmail.com",
			"name":  "Angular Talents",
		},
		"to": []interface{}{
			map[string]string{"email": receiverEmail},
		},
		"subject": subject,
		"text":    text,
		"html":    html,
	}

	client := &http.Client{}
	url := "https://send.api.mailtrap.io/api/send"

	internal.LogInfo("Preparing fallback email request", map[string]interface{}{
		"api_url":   url,
		"recipient": receiverEmail,
		"subject":   subject,
	})

	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		internal.LogInfo("Failed to marshal fallback email request", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		internal.LogInfo("Failed to create fallback email HTTP request", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Get token sample for logging
	tokenDebug := "not set"
	if mailTrapToken != "" {
		if len(mailTrapToken) > 8 {
			tokenDebug = mailTrapToken[:4] + "..." + mailTrapToken[len(mailTrapToken)-4:]
		} else {
			tokenDebug = mailTrapToken[:2] + "..."
		}
	}

	authHeader := "Bearer " + mailTrapToken
	authDebug := "Bearer " + tokenDebug
	internal.LogInfo("Setting fallback email authorization header", map[string]interface{}{
		"auth_header_sample": authDebug,
	})

	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/json")

	internal.LogInfo("Sending fallback email request", nil)
	resp, err := client.Do(req)
	if err != nil {
		internal.LogInfo("Failed to send fallback email request", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	defer resp.Body.Close()
	internal.LogInfo("Received response for fallback email", map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
	})

	if resp.StatusCode != http.StatusOK {
		// Read and log response body for debugging
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		internal.LogInfo("Mailtrap API error response for fallback email", map[string]interface{}{
			"response_body": bodyString,
		})

		// Create a new buffer with the same content for ReturnRawData
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		err := internal.NewError(http.StatusInternalServerError, "email.send_fallback_verification", "request to mailtrap api failed", "failed to send fallback verification email")
		internal.LogError(err, internal.ReturnRawData(resp))
		return err
	}

	internal.LogInfo("Successfully sent fallback verification email", map[string]interface{}{"email": receiverEmail})
	return nil
}
