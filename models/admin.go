package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// Admin represents an administrator user
type Admin struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username      string             `json:"username" bson:"username" validate:"required"`
	Password      string             `json:"password,omitempty" bson:"password" validate:"required"`
	Email         string             `json:"email" bson:"email" validate:"required,email"`
	FirstName     string             `json:"firstName" bson:"firstName"`
	LastName      string             `json:"lastName" bson:"lastName"`
	IsSuper       bool               `json:"isSuper" bson:"isSuper" default:"false"`
	AdminVerified bool               `json:"adminVerified" bson:"adminVerified" default:"false"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
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

// AdminLoginRequest represents admin login credentials
type AdminLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// NewAdminResponse creates a new AdminResponse from an Admin
func NewAdminResponse(admin *Admin) AdminResponse {
	return AdminResponse{
		ID:            admin.ID,
		Username:      admin.Username,
		Email:         admin.Email,
		FirstName:     admin.FirstName,
		LastName:      admin.LastName,
		IsSuper:       admin.IsSuper,
		AdminVerified: admin.AdminVerified,
		CreatedAt:     admin.CreatedAt,
		UpdatedAt:     admin.UpdatedAt,
	}
}

// HashPassword encrypts the admin's password
func (admin *Admin) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	admin.Password = string(hashedPassword)
	return nil
}

// ComparePassword checks if the provided password matches the stored hashed password
func (admin *Admin) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
}
