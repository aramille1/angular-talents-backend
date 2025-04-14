package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"angular-talents-backend/internal"
)


type FromStruct struct {
	Email string 	`json:"email"`
	Name string 	`json:"name"`
}
type Email struct {
	From FromStruct 	`json:"from"`

}

func HandleEmail(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	internal.LogInfo("TestEmail", map[string]interface{}{"emailtest": "OK"})

	requestBody := map[string]interface{}{
		"from": map[string]string{
			"email": "hello@angulartalents.com",
			"name": "Mailtrap Test",
		},
		"to": []interface{}{
			map[string]string{"email": "martin.axe@live.fr"},
		},
		"template_uuid": "269c54d2-872f-45be-aa5f-96584c1e40cd",
		"template_variables": map[string]string{
				"user_name": "Bob",
			},
	}

	client := &http.Client {}
	url := "https://send.api.mailtrap.io/api/send"

	marshalledRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println(err)
		return internal.NewError(http.StatusBadRequest, "err1", "failed to test email", err.Error())
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledRequestBody))
	if err != nil {
		fmt.Println(err)
		return internal.NewError(http.StatusBadRequest, "err2", "failed to test email", err.Error())
	}

	mailTrapToken := os.Getenv("MAILTRAP_TOKEN")
	req.Header.Add("Authorization", "Bearer "+mailTrapToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return internal.NewError(http.StatusBadRequest, "err3", "failed to test email", err.Error())
	}

	defer resp.Body.Close()
	var newBaseResponse interface{}
	err = json.NewDecoder(resp.Body).Decode(&newBaseResponse)
	if err != nil {
		fmt.Println(err)
		return internal.NewError(http.StatusBadRequest, "err4", "failed to test email", err.Error())
	}

	fmt.Println(resp.Status)
	fmt.Println(newBaseResponse)

	internal.LogInfo("Successfully sent test email", map[string]interface{}{"emailtest": "OK"})
	w.WriteResponse(http.StatusOK, map[string]interface{}{"res": "OK"})
	return nil
}
