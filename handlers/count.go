package handlers

import (
	"fmt"
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/internal"
)

func HandleCount(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	count, err:= dao.CountEngineers(r.Context())
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "login", "failed to login", "failed to authenticate")
	}

	w.WriteResponse(http.StatusOK, map[string]string{"engineers_count": fmt.Sprint(count)})
	return nil
}