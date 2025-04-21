package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"reverse-job-board/internal"
	"time"
)

// Fallback minimal countries list in case the external API is down
var fallbackCountries = []map[string]interface{}{
	{
		"name": map[string]interface{}{
			"common": "United States",
		},
		"flags": map[string]interface{}{
			"svg": "https://flagcdn.com/us.svg",
		},
	},
	{
		"name": map[string]interface{}{
			"common": "United Kingdom",
		},
		"flags": map[string]interface{}{
			"svg": "https://flagcdn.com/gb.svg",
		},
	},
	{
		"name": map[string]interface{}{
			"common": "Canada",
		},
		"flags": map[string]interface{}{
			"svg": "https://flagcdn.com/ca.svg",
		},
	},
	{
		"name": map[string]interface{}{
			"common": "Germany",
		},
		"flags": map[string]interface{}{
			"svg": "https://flagcdn.com/de.svg",
		},
	},
	{
		"name": map[string]interface{}{
			"common": "France",
		},
		"flags": map[string]interface{}{
			"svg": "https://flagcdn.com/fr.svg",
		},
	},
	{
		"name": map[string]interface{}{
			"common": "Australia",
		},
		"flags": map[string]interface{}{
			"svg": "https://flagcdn.com/au.svg",
		},
	},
	{
		"name": map[string]interface{}{
			"common": "India",
		},
		"flags": map[string]interface{}{
			"svg": "https://flagcdn.com/in.svg",
		},
	},
}

func HandleCountries(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	internal.LogInfo("Starting countries request", nil)

	// Set a timeout for the external API request
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make request to the restcountries API
	resp, err := client.Get("https://restcountries.com/v3.1/all?fields=name,flags")
	if err != nil {
		internal.LogInfo("External API request failed, using fallback data", map[string]interface{}{"error": err.Error()})
		return serveFallbackCountries(w)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		internal.LogInfo("External API returned non-200 status code", map[string]interface{}{"status_code": resp.StatusCode})
		return serveFallbackCountries(w)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		internal.LogInfo("Failed to read response body, using fallback data", map[string]interface{}{"error": err.Error()})
		return serveFallbackCountries(w)
	}

	// Set content type header to match the response from restcountries
	w.Header().Set("Content-Type", "application/json")

	// Write the raw JSON response directly to the client
	w.Write(body)

	internal.LogInfo("Successfully served countries data from external API", nil)
	return nil
}

func serveFallbackCountries(w internal.EnhancedResponseWriter) *internal.CustomError {
	// Convert fallback data to JSON
	jsonData, err := json.Marshal(fallbackCountries)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "countries.marshal_fallback", "failed to marshal fallback countries data", err.Error())
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.Write(jsonData)

	internal.LogInfo("Successfully served fallback countries data", nil)
	return nil
}
