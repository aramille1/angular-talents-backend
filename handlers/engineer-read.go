package handlers

import (
	"fmt"
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/gorilla/mux"
)

func HandleEngineerRead(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	params := mux.Vars(r.Request)
	engineerID := params["engineerID"]
	isMember := r.Context().Value("isMember").(bool)

	internal.LogInfo("Starting engineer read", map[string]interface{}{"user_id": r.Context().Value("userID"), "engineer_id": engineerID })

	if engineerID == "" || len(engineerID) != 36{
		fmt.Println("invalid engineerId param")
		return internal.NewError(http.StatusBadRequest, "engineer.read.validate_params", "failed to read engineer", "invalid engineerId param")
	}

	engineer, err := dao.FindEngineerById(r.Context(), engineerID)
	if err != nil {
		fmt.Println(err)
		return internal.NewError(http.StatusInternalServerError, "engineer.read.read_by_id", "failed to read engineer", err.Error())
	}
	
	if engineer == nil {
		return internal.NewError(http.StatusNotFound, "engineer.read.read_by_id", "failed to read engineer", "engineer not found")
	}

	if !isMember {
		partialEngineer := engineer.NewPartialEngineer()
		internal.LogInfo("Successfully read engineer partially concealed", map[string]interface{}{"user_id": r.Context().Value("userID"), "engineer_id": engineerID })
		w.WriteResponse(http.StatusOK,  map[string]domain.PartialEngineer{"engineer": *partialEngineer})	
		return nil
	}

	internal.LogInfo("Successfully read engineer", map[string]interface{}{"user_id": r.Context().Value("userID"), "engineer_id": engineerID })
	w.WriteResponse(http.StatusOK,  map[string]domain.Engineer{"engineer": *engineer})	
	return nil
}