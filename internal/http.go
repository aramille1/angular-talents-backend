package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ReturnRawData(response *http.Response) (map[string]interface{}) {
	var raw json.RawMessage
	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		fmt.Println("Error decoding response:", err)
	}

	// Use the data by decoding it further
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		fmt.Println("Error decoding raw message:", err)
	}

	return data
}
