package controllers

import (
	"net/http"

	"angular-talents-backend/models"
	"angular-talents-backend/services"

	"github.com/gin-gonic/gin"
)

// AdminController handles admin-related HTTP requests
type AdminController struct {
	adminService *services.AdminService
}

// NewAdminController creates a new instance of AdminController
func NewAdminController(adminService *services.AdminService) *AdminController {
	return &AdminController{
		adminService: adminService,
	}
}

// Login handles admin login requests
// @Summary Admin login
// @Description Authenticate an admin user and return a JWT token
// @Tags admins
// @Accept json
// @Produce json
// @Param credentials body models.AdminLoginRequest true "Admin credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /admin/login [post]
func (c *AdminController) Login(ctx *gin.Context) {
	var req models.AdminLoginRequest

	// Bind JSON request body to the model
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Authenticate admin
	admin, token, err := c.adminService.Login(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Return admin data and token
	ctx.JSON(http.StatusOK, gin.H{
		"admin": admin,
		"token": token,
	})
}

// GetProfile retrieves the admin profile for the authenticated admin
// @Summary Get admin profile
// @Description Get the profile of the authenticated admin
// @Tags admins
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.AdminResponse
// @Failure 401 {object} map[string]string
// @Router /admin/profile [get]
func (c *AdminController) GetProfile(ctx *gin.Context) {
	// Get admin ID from context (set by auth middleware)
	adminID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get admin profile
	admin, err := c.adminService.GetByID(ctx.Request.Context(), adminID.(string))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, admin)
}

// CreateAdmin creates a new admin
// @Summary Create admin
// @Description Create a new admin user (requires super-admin privileges)
// @Tags admins
// @Accept json
// @Produce json
// @Param admin body models.Admin true "Admin data"
// @Security ApiKeyAuth
// @Success 201 {object} models.AdminResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin [post]
func (c *AdminController) CreateAdmin(ctx *gin.Context) {
	var admin models.Admin

	// Bind JSON request body to the model
	if err := ctx.ShouldBindJSON(&admin); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Create admin
	createdAdmin, err := c.adminService.Create(ctx.Request.Context(), &admin)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, createdAdmin)
}

// GetAllAdmins retrieves all admins
// @Summary Get all admins
// @Description Get a list of all admin users (requires super-admin privileges)
// @Tags admins
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.AdminResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin [get]
func (c *AdminController) GetAllAdmins(ctx *gin.Context) {
	// Get all admins
	admins, err := c.adminService.GetAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, admins)
}

// GetAdmin retrieves an admin by ID
// @Summary Get admin by ID
// @Description Get an admin user by ID (requires super-admin privileges)
// @Tags admins
// @Produce json
// @Param id path string true "Admin ID"
// @Security ApiKeyAuth
// @Success 200 {object} models.AdminResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/{id} [get]
func (c *AdminController) GetAdmin(ctx *gin.Context) {
	id := ctx.Param("id")

	// Get admin by ID
	admin, err := c.adminService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, admin)
}

// UpdateAdmin updates an admin
// @Summary Update admin
// @Description Update an admin user (requires super-admin privileges or self-update)
// @Tags admins
// @Accept json
// @Produce json
// @Param id path string true "Admin ID"
// @Param admin body models.Admin true "Updated admin data"
// @Security ApiKeyAuth
// @Success 200 {object} models.AdminResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/{id} [put]
func (c *AdminController) UpdateAdmin(ctx *gin.Context) {
	id := ctx.Param("id")
	var updates models.Admin

	// Bind JSON request body to the model
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update admin
	updatedAdmin, err := c.adminService.Update(ctx.Request.Context(), id, &updates)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedAdmin)
}

// DeleteAdmin deletes an admin
// @Summary Delete admin
// @Description Delete an admin user (requires super-admin privileges)
// @Tags admins
// @Param id path string true "Admin ID"
// @Security ApiKeyAuth
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/{id} [delete]
func (c *AdminController) DeleteAdmin(ctx *gin.Context) {
	id := ctx.Param("id")

	// Delete admin
	if err := c.adminService.Delete(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
