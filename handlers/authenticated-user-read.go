package handlers

import (
	"fmt"
	"net/http"
	"angular-talents-backend/dao"
	"angular-talents-backend/internal"

	"github.com/google/uuid"
)

func HandleAuthenticatedUserRead(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	userID := r.Context().Value("userID").(uuid.UUID)

	internal.LogInfo("Starting read authenticated user", map[string]interface{}{"user_id": r.Context().Value("userID")})

	engineer, err := dao.FindEngineerByUser(r.Context(), userID)
	if err != nil {
		fmt.Println(err)
		return internal.NewError(http.StatusInternalServerError, "authenticated_user.read.read_engineer", "failed to read authenticated user", err.Error())
	}

	if engineer != nil {
		internal.LogInfo("Successfully read authenticated user", map[string]interface{}{"user_id": r.Context().Value("userID"), "engineer_id": engineer.ID})
		w.WriteResponse(http.StatusOK,  map[string]interface{}{"type": "engineer", "user": *engineer})
		return nil
	}
	
	recruiter, err := dao.FindRecruiterByUser(r.Context(), userID)
	if err != nil {
		fmt.Println(err)
		return internal.NewError(http.StatusInternalServerError, "authenticated_user.read.read_recruiter", "failed to read authenticated user", err.Error())
	}

	if recruiter != nil {
		internal.LogInfo("Successfully read authenticated user", map[string]interface{}{"user_id": r.Context().Value("userID"), "recruiter_id": recruiter.ID})
		w.WriteResponse(http.StatusOK,  map[string]interface{}{"type": "recruiter", "user": *recruiter})	
		return nil
	}


	return internal.NewError(http.StatusNotFound, "authenticated_user.read", "failed to read authenticated user", "could not find engineer or recruiter attached to user")
}