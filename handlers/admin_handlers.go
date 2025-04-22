package handlers

import (
	"context"
	"net/http"
	"reverse-job-board/db"
	"reverse-job-board/domain"
	"reverse-job-board/internal"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// AdminLoginRequest represents the login request for admin
type AdminLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// AdminResponse is the response structure for admin data (without sensitive information)
type AdminResponse struct {
	ID            primitive.ObjectID `json:"id,omitempty"`
	Username      string             `json:"username"`
	Email         string             `json:"email"`
	FirstName     string             `json:"firstName"`
	LastName      string             `json:"lastName"`
	IsSuper       bool               `json:"isSuper"`
	AdminVerified bool               `json:"adminVerified"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}

// HandleAdminLogin handles admin login requests
func HandleAdminLogin(w internal.EnhancedResponseWriter, r *internal.EnhancedRequest) *internal.CustomError {
	var req AdminLoginRequest
	if err := r.DecodeJSON(&w, &req); err != nil {
		return internal.NewError(
			http.StatusBadRequest,
			"admin.login.invalid_request",
			"Invalid request format",
			err.Error(),
		)
	}

	// Find admin by username
	collection := db.Database.Collection("admin")
	var admin struct {
		ID            primitive.ObjectID `bson:"_id,omitempty"`
		Username      string             `bson:"username"`
		Password      string             `bson:"password"`
		Email         string             `bson:"email"`
		FirstName     string             `bson:"firstName"`
		LastName      string             `bson:"lastName"`
		IsSuper       bool               `bson:"isSuper"`
		AdminVerified bool               `bson:"adminVerified"`
		CreatedAt     time.Time          `bson:"createdAt"`
		UpdatedAt     time.Time          `bson:"updatedAt"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"username": req.Username}).Decode(&admin)
	if err != nil {
		return internal.NewError(
			http.StatusUnauthorized,
			"admin.login.invalid_credentials",
			"Invalid credentials",
			"Admin not found",
		)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return internal.NewError(
			http.StatusUnauthorized,
			"admin.login.invalid_credentials",
			"Invalid credentials",
			"Password mismatch",
		)
	}

	// Generate JWT token
	token, err := domain.GenerateAdminJWT(admin.ID.Hex())
	if err != nil {
		return internal.NewError(
			http.StatusInternalServerError,
			"admin.login.token_generation",
			"Failed to generate token",
			err.Error(),
		)
	}

	// Create response
	response := struct {
		Admin AdminResponse `json:"admin"`
		Token string        `json:"token"`
	}{
		Admin: AdminResponse{
			ID:            admin.ID,
			Username:      admin.Username,
			Email:         admin.Email,
			FirstName:     admin.FirstName,
			LastName:      admin.LastName,
			IsSuper:       admin.IsSuper,
			AdminVerified: admin.AdminVerified,
			CreatedAt:     admin.CreatedAt,
			UpdatedAt:     admin.UpdatedAt,
		},
		Token: token,
	}

	w.WriteResponse(http.StatusOK, response)
	return nil
}
