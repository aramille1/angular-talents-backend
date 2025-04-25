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
	requestBody := map[string]interface{}{
		"from": map[string]string{
			"email": "hello@angulartalents.com",
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
			"email": "hello@angulartalents.com",
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
			"email": "hello@angulartalents.com",
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
