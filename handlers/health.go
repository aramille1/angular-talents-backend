package handlers

import (
	"net/http"
	"reverse-job-board/internal"
)
func HandleHealth(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	internal.LogInfo("Health", map[string]interface{}{"health": "OK"})
	w.WriteResponse(http.StatusOK, map[string]interface{}{"res": "OK"})
	return nil
}