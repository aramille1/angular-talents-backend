package handlers

import (
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func HandleAuthenticatedEngineerUpdate(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	userID := r.Context().Value("userID").(uuid.UUID)
	var engPayload domain.UpdateEngineerPayload

	internal.LogInfo("Starting authenticated engineer update", map[string]interface{}{"user_id": userID})

	err := r.DecodeJSON(&w, &engPayload)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "authenticated_engineer.update.decode_body", "failed to update engineer", err.Error())
	}

	v := validator.New()
	err = v.Struct(engPayload)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "authenticated_engineer.udpate.validate", "failed to update engineer", err.Error())
	}

	updatedEng, err := dao.UpdateEngineerByUser(r.Context(), userID , &engPayload)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "authenticated_engineer.update.update_table", "failed to update engineer", err.Error())
	}

	internal.LogInfo("Successfully updated authenticated engineer", map[string]interface{}{"engineer": updatedEng})
	w.WriteResponse(http.StatusOK, map[string]*domain.Engineer{"engineer": updatedEng})
	return nil
}