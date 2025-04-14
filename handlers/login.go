package handlers

import (
	"net/http"
	"angular-talents-backend/domain"
	"angular-talents-backend/internal"

	"github.com/go-playground/validator/v10"
)

func HandleLogin(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	var lg domain.LoginData
	err := r.DecodeJSON(&w, &lg)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "login.decode_body", "failed to login", err.Error())
	}
	
	v := validator.New()
	err = v.Struct(lg)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "login.validate_body", "failed to login", err.Error())
	}

	authenticatedUser, err := lg.VerifyLogin(r.Context())
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "login", "failed to login", "failed to authenticate")
	}

	tokenString, err:= domain.GenerateJWT(authenticatedUser.ID)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "login", "failed to login", "failed to authenticate")
	}

	internal.LogInfo("Successfully loggedin user", map[string]interface{}{"user_id": authenticatedUser.ID })
	w.WriteResponse(http.StatusOK, map[string]string{"auth_token": tokenString})
	return nil
}