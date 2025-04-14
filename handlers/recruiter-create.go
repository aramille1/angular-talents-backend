package handlers

import (
	"net/http"
	"angular-talents-backend/dao"
	"angular-talents-backend/domain"
	"angular-talents-backend/internal"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)


func HandleRecruiterCreate(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	var recruiterPayload domain.CreateRecruiterPayload
	internal.LogInfo("Starting recruiter creation for user", map[string]interface{}{"user_id": r.Context().Value("userID") })

	err := r.DecodeJSON(&w, &recruiterPayload)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "recruiter.create.decode_body", "failed to create new recruiter", err.Error())
	}

	v := validator.New()
	err = v.Struct(recruiterPayload)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "recruiter.create.validate_body", "failed to create new recruiter", err.Error())
	}

	recruiter, err := recruiterPayload.NewRecruiter(r.Context())
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "recruiter.create.create_new_recruiter", "failed to create new recruiter", err.Error())
	}

	err = internal.Validate(r.Context(), recruiter.ID)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "recruiter.create.validate_new_recruiter", "failed to create new recruiter", err.Error())
	}

	_, err = dao.InsertNewRecruiter(r.Context(), recruiter)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "recruiter.create.insert", "failed to create new recruiter", err.Error())
	}

	internal.LogInfo("Successfully created new recruiter", map[string]interface{}{"recruiterId": recruiter.ID})
	w.WriteResponse(http.StatusOK, map[string]uuid.UUID{"recruiterId": recruiter.ID})
	return nil
}