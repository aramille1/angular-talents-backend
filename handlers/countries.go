package handlers

import (
	"io/ioutil"
	"net/http"
	"reverse-job-board/internal"
)

func HandleCountries(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	internal.LogInfo("Starting countries request", nil)

	// Make request to the restcountries API
	resp, err := http.Get("https://restcountries.com/v3.1/all?fields=name,flags")
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "countries.request_failed", "failed to fetch countries", err.Error())
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "countries.read_response", "failed to read countries response", err.Error())
	}

	// Set content type header to match the response from restcountries
	w.Header().Set("Content-Type", "application/json")

	// Write the raw JSON response directly to the client
	w.Write(body)

	internal.LogInfo("Successfully served countries data", nil)
	return nil
}
