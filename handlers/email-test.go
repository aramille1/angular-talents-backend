package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reverse-job-board/internal"
)

type FromStruct struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
type Email struct {
	From FromStruct `json:"from"`
}

func HandleEmail(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	internal.LogInfo("Starting test email request", map[string]interface{}{"test": "email"})

	// Get token for diagnostic purposes
	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")
	tokenDebug := "not set"
	if mailTrapToken != "" {
		if len(mailTrapToken) > 8 {
			tokenDebug = mailTrapToken[:4] + "..." + mailTrapToken[len(mailTrapToken)-4:]
		} else {
			tokenDebug = mailTrapToken[:2] + "..."
		}
	}

	internal.LogInfo("Test email configuration", map[string]interface{}{
		"token_sample": tokenDebug,
		"token_length": len(mailTrapToken),
	})

	requestBody := map[string]interface{}{
		"from": map[string]string{
			"email": "hello.angulartalents@gmail.com",
			"name":  "Mailtrap Test",
		},
		"to": []interface{}{
			map[string]string{"email": "test@example.com"},
		},
		"subject": "API Test - Angular Talents",
		"text":    "This is a test email from Angular Talents to verify the Mailtrap API is working.",
		"html":    "<h1>Test Email</h1><p>This is a test email to verify the Mailtrap API is working.</p>",
	}

	client := &http.Client{}
	url := "https://send.api.mailtrap.io/api/send"

	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		internal.LogInfo("Failed to marshal test email request body", map[string]interface{}{"error": err.Error()})
		return internal.NewError(http.StatusBadRequest, "test.email.marshal", "Failed to marshal test email request", err.Error())
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		internal.LogInfo("Failed to create test email HTTP request", map[string]interface{}{"error": err.Error()})
		return internal.NewError(http.StatusBadRequest, "test.email.create_request", "Failed to create test email request", err.Error())
	}

	authHeader := "Bearer " + mailTrapToken
	authDebug := "Bearer " + tokenDebug
	internal.LogInfo("Setting test email authorization header", map[string]interface{}{
		"auth_header_sample": authDebug,
	})

	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/json")

	internal.LogInfo("Sending test email request", nil)
	resp, err := client.Do(req)
	if err != nil {
		internal.LogInfo("Failed to send test email", map[string]interface{}{"error": err.Error()})
		return internal.NewError(http.StatusBadRequest, "test.email.send", "Failed to send test email", err.Error())
	}

	defer resp.Body.Close()

	internal.LogInfo("Received test email response", map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
	})

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		internal.LogInfo("Failed to read test email response body", map[string]interface{}{"error": err.Error()})
		return internal.NewError(http.StatusBadRequest, "test.email.read_response", "Failed to read test email response", err.Error())
	}

	bodyString := string(bodyBytes)
	internal.LogInfo("Test email response body", map[string]interface{}{
		"response_body": bodyString,
	})

	// Parse the response for better diagnostics
	var responseData map[string]interface{}
	err = json.Unmarshal(bodyBytes, &responseData)
	if err != nil {
		internal.LogInfo("Failed to parse test email response JSON", map[string]interface{}{
			"error":        err.Error(),
			"raw_response": bodyString,
		})
	}

	// Recreate response body for the next reader
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Check if successful
	if resp.StatusCode == http.StatusOK {
		internal.LogInfo("Test email sent successfully", nil)
		w.WriteResponse(http.StatusOK, map[string]interface{}{
			"status":   "success",
			"message":  "Test email sent successfully",
			"response": responseData,
		})
		return nil
	} else {
		errorMsg := "Failed to send test email"
		if responseData != nil {
			if errors, ok := responseData["errors"].([]interface{}); ok && len(errors) > 0 {
				errorMsg = fmt.Sprintf("Mailtrap API error: %v", errors[0])
			}
		}

		internal.LogInfo("Test email failed", map[string]interface{}{
			"error_message": errorMsg,
			"response":      responseData,
		})

		return internal.NewError(resp.StatusCode, "test.email.failed", errorMsg, bodyString)
	}
}
