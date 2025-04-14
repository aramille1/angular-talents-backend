package handlers

import (
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func HandleRecruiterUpdate(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	userID := r.Context().Value("userID").(uuid.UUID)
	var recruiterPayload domain.UpdateRecruiterPayload
	params := mux.Vars(r.Request)
	recruiterID := params["recruiterID"]
	internal.LogInfo("Starting recruiter update", map[string]interface{}{"user_id": userID, "recruiter_id": recruiterID })

	err := r.DecodeJSON(&w, &recruiterPayload)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "recruiter.update.decode_body", "failed to update recruiter", err.Error())
	}

	err = recruiterPayload.Validate(r.Context(), userID, recruiterID)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "recruiter.udpate.validate", "failed to update recruiter", err.Error())
	}

	udpatedRecruiter, err := dao.UpdateRecruiter(r.Context(), recruiterID , &recruiterPayload)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "recruiter.update.update_table", "failed to update recruiter", err.Error())
	}

	internal.LogInfo("Successfully updated recruiter", map[string]interface{}{"recruiter": udpatedRecruiter})
	w.WriteResponse(http.StatusOK, map[string]*domain.Recruiter{"recruiter": udpatedRecruiter})
	return nil
}