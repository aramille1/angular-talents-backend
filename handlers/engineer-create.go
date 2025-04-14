package handlers

import (
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func HandleEngineerCreate(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	var engPayload domain.CreateEngineerPayload
	internal.LogInfo("Starting engineer creation for user", map[string]interface{}{"user_id": r.Context().Value("userID") })

	err := r.DecodeJSON(&w, &engPayload)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "engineer.create.decode_body", "failed to create new engineer", err.Error())
	}

	v := validator.New()
	err = v.Struct(engPayload)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "engineer.create.validate_body", "failed to create new engineer", err.Error())
	}

	eng, err := engPayload.NewEngineer(r.Context())
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "engineer.create.create_new_engineer", "failed to create new engineer", err.Error())
	}

	err = internal.Validate(r.Context(), eng.ID)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "engineer.create.validate_new_engineer", "failed to create new engineer", err.Error())
	}

	_, err = dao.InsertNewEngineer(r.Context(), eng)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "engineer.create.insert", "failed to create new engineer", err.Error())
	}

	internal.LogInfo("Successfully created new engineer", map[string]interface{}{"engineerId": eng.ID})
	w.WriteResponse(http.StatusOK, map[string]uuid.UUID{"engineerId": eng.ID})
	return nil
}