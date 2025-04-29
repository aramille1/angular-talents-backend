package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// SlackWebhookMessage represents the structure of a Slack message
type SlackWebhookMessage struct {
	Text        string            `json:"text,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Footer    string       `json:"footer,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

// SendSlackNotification sends a notification to Slack via webhook
func SendSlackNotification(webhookURL string, message SlackWebhookMessage) error {
	if webhookURL == "" {
		LogInfo("Slack webhook URL not set. Skipping notification.", nil)
		return nil
	}

	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned %d status code", resp.StatusCode)
	}

	LogInfo("Slack notification sent successfully", map[string]interface{}{
		"status_code": resp.StatusCode,
	})
	return nil
}

// NotifyNewUserRegistration sends a notification when a new user registers
func NotifyNewUserRegistration(email string, userID string) {
	webhookURL := os.Getenv("SLACK_NEW_USER_WEBHOOK_URL")

	message := SlackWebhookMessage{
		Attachments: []SlackAttachment{
			{
				Color:     "#36a64f", // Green
				Title:     "🎉 New User Registration",
				Text:      "A new user has registered on Angular Talents",
				Timestamp: time.Now().Unix(),
				Fields: []SlackField{
					{
						Title: "Email",
						Value: email,
						Short: true,
					},
					{
						Title: "User ID",
						Value: userID,
						Short: true,
					},
				},
				Footer: "Angular Talents",
			},
		},
	}

	if err := SendSlackNotification(webhookURL, message); err != nil {
		LogInfo("Failed to send Slack notification for new user", map[string]interface{}{
			"error": err.Error(),
			"email": email,
		})
	}
}

// NotifyNewEngineerProfile sends a notification when a new engineer profile is created
func NotifyNewEngineerProfile(firstName, lastName string, engineerID string) {
	webhookURL := os.Getenv("SLACK_NEW_ENGINEER_WEBHOOK_URL")

	message := SlackWebhookMessage{
		Attachments: []SlackAttachment{
			{
				Color:     "#3498db", // Blue
				Title:     "👨‍💻 New Engineer Profile",
				Text:      "A new engineer profile has been created on Angular Talents",
				Timestamp: time.Now().Unix(),
				Fields: []SlackField{
					{
						Title: "Name",
						Value: firstName + " " + lastName,
						Short: true,
					},
					{
						Title: "Engineer ID",
						Value: engineerID,
						Short: true,
					},
				},
				Footer: "Angular Talents",
			},
		},
	}

	if err := SendSlackNotification(webhookURL, message); err != nil {
		LogInfo("Failed to send Slack notification for new engineer", map[string]interface{}{
			"error": err.Error(),
			"name":  firstName + " " + lastName,
		})
	}
}

// NotifyNewRecruiterProfile sends a notification when a new recruiter profile is created
func NotifyNewRecruiterProfile(firstName, lastName, company string, recruiterID string) {
	webhookURL := os.Getenv("SLACK_NEW_RECRUITER_WEBHOOK_URL")

	message := SlackWebhookMessage{
		Attachments: []SlackAttachment{
			{
				Color:     "#e74c3c", // Red
				Title:     "🏢 New Recruiter Profile",
				Text:      "A new recruiter profile has been created on Angular Talents",
				Timestamp: time.Now().Unix(),
				Fields: []SlackField{
					{
						Title: "Name",
						Value: firstName + " " + lastName,
						Short: true,
					},
					{
						Title: "Company",
						Value: company,
						Short: true,
					},
					{
						Title: "Recruiter ID",
						Value: recruiterID,
						Short: false,
					},
				},
				Footer: "Angular Talents",
			},
		},
	}

	if err := SendSlackNotification(webhookURL, message); err != nil {
		LogInfo("Failed to send Slack notification for new recruiter", map[string]interface{}{
			"error":   err.Error(),
			"name":    firstName + " " + lastName,
			"company": company,
		})
	}
}
