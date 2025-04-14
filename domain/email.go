
package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"angular-talents-backend/internal"
)


func SendNewEmail(templateId, userId, receiverEmail string, code int) error {
	internal.LogInfo("Starting sending sign up confirmation email", map[string]interface{}{"user_id": userId })
	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")
		requestBody := map[string]interface{}{
		"from": map[string]string{
			"email": "hello@angulartalents.com",
			"name": "The Angular Team",
		},
		"to": []interface{}{
			map[string]string{"email": receiverEmail},
		},
		"template_uuid": templateId,
		"template_variables": map[string]string{
			"user_id": userId,
			"verification_code": fmt.Sprint(code),
		},
	}

	client := &http.Client {}
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

	internal.LogInfo("Successfully sent sign up confirmation email", map[string]interface{}{"user_id": userId })
	return nil
}
