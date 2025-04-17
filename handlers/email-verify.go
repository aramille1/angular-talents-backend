package handlers

import (
	"fmt"
	"net/http"
	"reverse-job-board/dao"
	"reverse-job-board/domain"
	"reverse-job-board/internal"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func HandleEmailVerify(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	params := mux.Vars(r.Request)
	userId := params["userID"]
	verificationCode := params["verificationCode"]
	internal.LogInfo("Starting email verification", map[string]interface{}{"user_id": r.Context().Value("userID")})

	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "email.verify.parse_user_id", "failed to verify email", err.Error())
	}

	user, err := dao.FindUserById(r.Context(), parsedUserId)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "email.verify.read_user_by_id", "failed to verify email", err.Error())
	}

	if user == nil {
		return internal.NewError(http.StatusInternalServerError, "email.verify.user_not_found", "failed to verify email", "user not found")
	}

	if fmt.Sprint(user.VerificationCode) != verificationCode {
		return internal.NewError(http.StatusBadRequest, "email.verify.check_verification_code", "failed to verify email", "verification code incorrect")
	}

	err = dao.UpdateUserVerifiedStatus(r.Context(), userId)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "email.verify.update_user_verified_status", "failed to verify email", err.Error())
	}

	internal.LogInfo("Successfully verified email", map[string]interface{}{"user_id": r.Context().Value("userID")})
	w.WriteResponse(http.StatusOK, map[string]domain.User{"user": *user})
	return nil
}
