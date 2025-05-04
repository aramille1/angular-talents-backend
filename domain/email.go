package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reverse-job-board/internal"
)

func SendNewEmail(templateId, userId, receiverEmail string, code int) error {
	internal.LogInfo("Starting sending sign up confirmation email", map[string]interface{}{"user_id": userId})
	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")

	// Check if Mailtrap is configured
	if mailTrapToken == "" || templateId == "" {
		internal.LogInfo("Mailtrap not configured, using fallback email verification method", nil)
		return logVerificationInstructions(userId, receiverEmail, code)
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

	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		internal.LogInfo("Failed to marshal Mailtrap request body, using fallback", nil)
		return logVerificationInstructions(userId, receiverEmail, code)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		internal.LogInfo("Failed to create Mailtrap request, using fallback", nil)
		return logVerificationInstructions(userId, receiverEmail, code)
	}

	req.Header.Add("Authorization", "Bearer "+mailTrapToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		internal.LogInfo("Failed to send request to Mailtrap, using fallback", nil)
		return logVerificationInstructions(userId, receiverEmail, code)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := internal.NewError(http.StatusInternalServerError, "email.send", "request to mailtrap api failed", "failed to send email")
		internal.LogError(err, internal.ReturnRawData(resp))
		// Use fallback method
		internal.LogInfo("Mailtrap API returned error, using fallback", map[string]interface{}{"status": resp.StatusCode})
		return logVerificationInstructions(userId, receiverEmail, code)
	}

	internal.LogInfo("Successfully sent sign up confirmation email", map[string]interface{}{"user_id": userId})
	return nil
}

// logVerificationInstructions logs verification instructions for admin to manually verify users
// This is a fallback method when email sending fails
func logVerificationInstructions(userId, email string, code int) error {
	verificationLink := fmt.Sprintf("%s/verify-email/%s/%d", os.Getenv("FRONTEND_URL"), userId, code)

	instruction := fmt.Sprintf(`
======================= MANUAL VERIFICATION REQUIRED =======================
User %s (%s) could not receive verification email.
Please provide them with this verification link manually:
%s
==========================================================================
`, email, userId, verificationLink)

	internal.LogInfo("Manual verification required", map[string]interface{}{
		"user_id":           userId,
		"email":             email,
		"verification_link": verificationLink,
	})

	// Create a "fallback" notification in Slack
	if os.Getenv("SLACK_WEBHOOK_URL") != "" {
		go internal.NotifyEmailFailure(email, userId, verificationLink)
	}

	fmt.Println(instruction)

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
