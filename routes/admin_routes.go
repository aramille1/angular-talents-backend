package routes

import (
	"angular-talents-backend/controllers"
	"angular-talents-backend/middleware"
	"reverse-job-board/services"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes registers all admin-related routes
func RegisterAdminRoutes(router *gin.Engine, adminController *controllers.AdminController, adminService *services.AdminService) {
	// Admin group
	adminGroup := router.Group("/api/admin")

	// Public routes
	adminGroup.POST("/login", adminController.Login)

	// Protected routes (require admin authentication)
	protected := adminGroup.Use(middleware.AdminAuthMiddleware(adminService))
	protected.GET("/profile", adminController.GetProfile)

	// Recruiter management
	protected.GET("/recruiters", adminController.GetRecruiters)
	protected.GET("/recruiters/pending", adminController.GetPendingRecruiters)
	protected.PATCH("/recruiters/:id/approve", adminController.ApproveRecruiter)
	protected.PATCH("/recruiters/:id/reject", adminController.RejectRecruiter)

	// Admin management (requires super admin privileges)
	superAdmin := protected.Use(middleware.SuperAdminAuthMiddleware())
	superAdmin.POST("", adminController.CreateAdmin)
	superAdmin.GET("", adminController.GetAllAdmins)
	superAdmin.GET("/:id", adminController.GetAdmin)

	// Admin update/delete (requires super admin privileges or self-modification)
	selfOrSuperAdmin := protected.Use(middleware.SelfOrSuperAdminAuthMiddleware())
	selfOrSuperAdmin.PUT("/:id", adminController.UpdateAdmin)
	selfOrSuperAdmin.DELETE("/:id", adminController.DeleteAdmin)
}
