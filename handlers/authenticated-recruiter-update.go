package handlers

import (
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func HandleAuthenticatedRecruiterUpdate(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	userID := r.Context().Value("userID").(uuid.UUID)
	var recruiterPayload domain.UpdateRecruiterPayload

	internal.LogInfo("Starting authenticated recruiter update", map[string]interface{}{"user_id": userID})

	err := r.DecodeJSON(&w, &recruiterPayload)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "authenticated_recruiter.update.decode_body", "failed to update recruiter", err.Error())
	}

	v := validator.New()
	err = v.Struct(recruiterPayload)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "authenticated_recruiter.udpate.validate", "failed to update recruiter", err.Error())
	}

	updatedRecruiter, err := dao.UpdateRecruiterByUser(r.Context(), userID , &recruiterPayload)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "authenticated_recruiter.update.update_table", "failed to update recruiter", err.Error())
	}

	internal.LogInfo("Successfully updated authenticated recruiter", map[string]interface{}{"recruiter": updatedRecruiter})
	w.WriteResponse(http.StatusOK, map[string]*domain.Recruiter{"recruiter": updatedRecruiter})
	return nil
}