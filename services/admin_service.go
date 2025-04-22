package services

import (
	"context"
	"errors"
	"time"

	"github.com/angular-talents/angular-talents-backend/models"
	"github.com/angular-talents/angular-talents-backend/repositories"
	"github.com/angular-talents/angular-talents-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminService handles business logic for admin operations
type AdminService struct {
	adminRepo *repositories.AdminRepository
}

// NewAdminService creates a new instance of AdminService
func NewAdminService(adminRepo *repositories.AdminRepository) *AdminService {
	return &AdminService{
		adminRepo: adminRepo,
	}
}

// Login authenticates an admin and generates a JWT token
func (s *AdminService) Login(ctx context.Context, req *models.AdminLoginRequest) (*models.AdminResponse, string, error) {
	// Find admin by username
	admin, err := s.adminRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Verify password
	if err := admin.ComparePassword(req.Password); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Generate JWT token
	claims := utils.JWTClaims{
		ID:        admin.ID.Hex(),
		Role:      "admin",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	token, err := utils.GenerateJWT(claims)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	// Return admin data and token
	adminResp := models.NewAdminResponse(admin)
	return &adminResp, token, nil
}

// GetByID retrieves an admin by ID
func (s *AdminService) GetByID(ctx context.Context, id string) (*models.AdminResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	admin, err := s.adminRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	adminResp := models.NewAdminResponse(admin)
	return &adminResp, nil
}

// Create creates a new admin
func (s *AdminService) Create(ctx context.Context, admin *models.Admin) (*models.AdminResponse, error) {
	// Check if username already exists
	existingAdmin, err := s.adminRepo.FindByUsername(ctx, admin.Username)
	if err == nil && existingAdmin != nil {
		return nil, errors.New("username already taken")
	}

	// Check if email already exists
	existingAdmin, err = s.adminRepo.FindByEmail(ctx, admin.Email)
	if err == nil && existingAdmin != nil {
		return nil, errors.New("email already in use")
	}

	// Hash password
	if err := admin.HashPassword(); err != nil {
		return nil, errors.New("error hashing password")
	}

	// Create admin
	createdAdmin, err := s.adminRepo.Create(ctx, admin)
	if err != nil {
		return nil, errors.New("error creating admin")
	}

	adminResp := models.NewAdminResponse(createdAdmin)
	return &adminResp, nil
}

// Update updates an existing admin
func (s *AdminService) Update(ctx context.Context, id string, updates *models.Admin) (*models.AdminResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	// Get the existing admin
	existingAdmin, err := s.adminRepo.FindByID(ctx, objID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	// Update fields
	if updates.Username != "" {
		existingAdmin.Username = updates.Username
	}
	if updates.Email != "" {
		existingAdmin.Email = updates.Email
	}
	if updates.FirstName != "" {
		existingAdmin.FirstName = updates.FirstName
	}
	if updates.LastName != "" {
		existingAdmin.LastName = updates.LastName
	}
	if updates.Password != "" {
		existingAdmin.Password = updates.Password
		if err := existingAdmin.HashPassword(); err != nil {
			return nil, errors.New("error hashing password")
		}
	}

	// Save updates
	if err := s.adminRepo.Update(ctx, existingAdmin); err != nil {
		return nil, errors.New("error updating admin")
	}

	adminResp := models.NewAdminResponse(existingAdmin)
	return &adminResp, nil
}

// Delete deletes an admin
func (s *AdminService) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID format")
	}

	return s.adminRepo.Delete(ctx, objID)
}

// GetAll retrieves all admins
func (s *AdminService) GetAll(ctx context.Context) ([]*models.AdminResponse, error) {
	admins, err := s.adminRepo.FindAll(ctx)
	if err != nil {
		return nil, errors.New("error retrieving admins")
	}

	// Convert to response objects
	var adminResponses []*models.AdminResponse
	for _, admin := range admins {
		resp := models.NewAdminResponse(admin)
		adminResponses = append(adminResponses, &resp)
	}

	return adminResponses, nil
}
