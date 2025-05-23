package handlers

import (
	"angular-talents-backend/dao"
	"angular-talents-backend/domain"
	"angular-talents-backend/internal"
	"net/http"
	"os"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)


func HandleSignUp(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError{
	internal.LogInfo("Starting sign up", nil)
	confirmEmailTemplateId := os.Getenv("CONFIRM_EMAIL_TEMPLATE_ID")
	var userData domain.SignUpData
	err := r.DecodeJSON(&w, &userData)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.decode_body", "failed to sign up", err.Error())
	}

	v := validator.New()
	err = v.Struct(userData)
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "signup.validate_body", "failed to sign up", err.Error())
	}

	user, err := userData.NewUser()
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.create_user", "failed to sign up", err.Error())
	}

	err = user.Validate(r.Context())
	if err != nil {
		return internal.NewError(http.StatusBadRequest, "signup.validate_user", "failed to sign up", err.Error())
	}

	userId, err := dao.InsertNewUser(r.Context(), user)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.insert_user", "failed to sign up", err.Error())
	}

	err = domain.SendNewEmail(confirmEmailTemplateId, userId, user.Email, user.VerificationCode)
	if err != nil {
		return internal.NewError(http.StatusInternalServerError, "signup.send_confirmation_email", "failed to sign up", err.Error())
	}

	internal.LogInfo("Successfully signed up user", map[string]interface{}{"user_id": user.ID })
	w.WriteResponse(http.StatusOK, map[string]uuid.UUID{"user_id": user.ID})
	return nil
}
