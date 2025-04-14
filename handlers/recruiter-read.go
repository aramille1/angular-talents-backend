package handlers

import (
	"fmt"
	"net/http"
	"angular-talents-backend/dao"
	"angular-talents-backend/domain"
	"angular-talents-backend/internal"

	"github.com/gorilla/mux"
)

func HandleRecruiterRead(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	params := mux.Vars(r.Request)
	recruiterID := params["recruiterID"]

	internal.LogInfo("Starting recruiter read", map[string]interface{}{"user_id": r.Context().Value("userID"), "recruiter_id": recruiterID })

	if recruiterID == "" || len(recruiterID) != 36{
		fmt.Println("invalid recruiterId param")
		return internal.NewError(http.StatusBadRequest, "recruiter.read.validate_params", "failed to read recruiter", "invalid recruiterId param")
	}

	recruiter, err := dao.FindRecruiterById(r.Context(), recruiterID)
	if err != nil {
		fmt.Println(err)
		return internal.NewError(http.StatusInternalServerError, "recruiter.read.read_by_id", "failed to read recruiter", err.Error())
	}
	
	if recruiter == nil {
		return internal.NewError(http.StatusNotFound, "recruiter.read.read_by_id", "failed to read recruiter", "recruiter not found")
	}

	internal.LogInfo("Successfully read recruiter", map[string]interface{}{"user_id": r.Context().Value("userID"), "recruiter_id": recruiterID })
	w.WriteResponse(http.StatusOK,  map[string]domain.Recruiter{"recruiter": *recruiter})	
	return nil
}