package handlers

import (
	"net/http"
	"angular-talents-backend/internal"
)
func HandleHealth(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	internal.LogInfo("Health", map[string]interface{}{"health": "OK"})
	w.WriteResponse(http.StatusOK, map[string]interface{}{"res": "OK"})
	return nil
}
