package handlers

import (
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func HandleEngineerUpdate(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	userID := r.Context().Value("userID").(uuid.UUID)
	var engPayload domain.UpdateEngineerPayload
	params := mux.Vars(r.Request)
	engineerID := params["engineerID"]
	internal.LogInfo("Starting engineer update", map[string]interface{}{"user_id": userID, "engineer_id": engineerID })

	err := r.DecodeJSON(&w, &engPayload)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "engineer.update.decode_body", "failed to update engineer", err.Error())
	}

	err = engPayload.Validate(r.Context(), userID, engineerID)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "engineer.udpate.validate", "failed to update engineer", err.Error())
	}

	updatedEng, err := dao.UpdateEngineer(r.Context(), engineerID , &engPayload)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "engineer.update.update_table", "failed to update engineer", err.Error())
	}

	internal.LogInfo("Successfully updated engineer", map[string]interface{}{"engineer": updatedEng})
	w.WriteResponse(http.StatusOK, map[string]*domain.Engineer{"engineer": updatedEng})
	return nil
}