package handlers

import (
	"net/http"
	"angular-talents-backend/dao"
	"angular-talents-backend/domain"
	"angular-talents-backend/internal"
)

func HandleEngineerList(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError{
	internal.LogInfo("Starting engineer list", map[string]interface{}{"user_id": r.Context().Value("userID")})
	isMember := r.Context().Value("isMember").(bool)

	listParams, err := domain.NewListEngineerParams(isMember, r.URL.Query())
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "engineer.list.new_query_params", "failed to list engineers", err.Error())
	}

	engineers, err := dao.ReadEngineers(r.Context(), listParams)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "engineer.list.read_engineers", "failed to list engineers", err.Error())
	}

	internal.LogInfo("Successfully listed engineers", map[string]interface{}{"user_id": r.Context().Value("userID")})
	w.WriteResponse(http.StatusOK,  map[string][]*domain.Engineer{"engineers": engineers})
	return nil
}