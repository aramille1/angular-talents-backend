package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// SlackMessage represents the payload to send to Slack
type SlackMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Slack message attachment with fields
type Attachment struct {
	Color     string  `json:"color"`
	Title     string  `json:"title,omitempty"`
	TitleLink string  `json:"title_link,omitempty"`
	Text      string  `json:"text,omitempty"`
	Fields    []Field `json:"fields,omitempty"`
	Footer    string  `json:"footer,omitempty"`
	Timestamp int64   `json:"ts,omitempty"`
}

// Field represents a field in a Slack attachment
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NotifyNewRecruiter sends a notification to Slack when a new business profile is created
func NotifyNewRecruiter(recruiterId string, companyName, firstName, lastName, email string) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		log.WithFields(log.Fields{}).Warn("Slack webhook URL not configured, skipping notification")
		return nil
	}

	// Generate the admin approval URL
	adminURL := os.Getenv("ADMIN_PANEL_URL")
	if adminURL == "" {
		adminURL = "http://localhost:4200/admin"
	}

	// Create the Slack message
	now := time.Now()
	formattedTime := now.Format("15:04 - 02 Jan 2006")

	message := SlackMessage{
		Text: fmt.Sprintf("🚨 *New Business Profile Created - Needs Approval - %s*", formattedTime),
		Attachments: []Attachment{
			{
				Color: "#36a64f",
				Title: fmt.Sprintf("New profile from %s", companyName),
				Text:  "A new business profile has been created and is awaiting approval.",
				Fields: []Field{
					{
						Title: "Company",
						Value: companyName,
						Short: true,
					},
					{
						Title: "Contact",
						Value: fmt.Sprintf("%s %s", firstName, lastName),
						Short: true,
					},
					{
						Title: "Email",
						Value: email,
						Short: true,
					},
					{
						Title: "ID",
						Value: recruiterId,
						Short: true,
					},
				},
				Footer:    "Angular Talents Platform",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	// Add action button text
	message.Attachments[0].Text += fmt.Sprintf("\n\n<<%s|Click to open Admin Panel and review>>", adminURL)

	// Convert message to JSON
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Error("Failed to marshal Slack message")
		return err
	}

	// Send the HTTP request to the webhook
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonMessage))
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Error("Failed to send Slack notification")
		return err
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{"status": resp.StatusCode}).Error("Failed to send Slack notification")
		return fmt.Errorf("failed to send Slack notification, status code: %d", resp.StatusCode)
	}

	LogInfo("Slack notification sent successfully", map[string]interface{}{"recruiterId": recruiterId})
	return nil
}
