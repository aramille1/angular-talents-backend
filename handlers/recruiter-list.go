package handlers

import (
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"
)

// HandleRecruiterList handles requests to get all recruiters
func HandleRecruiterList(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	internal.LogInfo("Starting recruiter list", map[string]interface{}{"user_id": r.Context().Value("userID")})

	// Fetch all recruiters from the database
	recruiters, err := dao.ReadAllRecruiters(r.Context())
	if err != nil {
		return internal.NewError(
			http.StatusInternalServerError,
			"recruiter.list.read_recruiters",
			"Failed to list recruiters",
			err.Error(),
		)
	}

	internal.LogInfo("Successfully listed recruiters", map[string]interface{}{"user_id": r.Context().Value("userID"), "count": len(recruiters)})

	// Return the recruiters as JSON
	w.WriteResponse(http.StatusOK, map[string][]*domain.Recruiter{"recruiters": recruiters})
	return nil
}
