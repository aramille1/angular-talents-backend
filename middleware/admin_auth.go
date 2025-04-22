package middleware

import (
	"net/http"
	"strings"

	"angular-talents-backend/models"
	"angular-talents-backend/services"
	"angular-talents-backend/utils"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware is a middleware that authenticates admin users
func AdminAuthMiddleware(adminService *services.AdminService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Check if the token belongs to an admin
		if claims.Role != "admin" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Not authorized to access admin resources"})
			return
		}

		// Verify that the admin exists
		admin, err := adminService.GetByID(ctx.Request.Context(), claims.ID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Admin not found"})
			return
		}

		// Set admin ID in context
		ctx.Set("userID", claims.ID)
		ctx.Set("userRole", claims.Role)
		ctx.Set("admin", admin)

		ctx.Next()
	}
}

// SuperAdminAuthMiddleware is a middleware that ensures the admin is a super admin
func SuperAdminAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get admin from context
		adminInterface, exists := ctx.Get("admin")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Check if admin is a super admin
		admin, ok := adminInterface.(*models.AdminResponse)
		if !ok || !admin.IsSuper {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Super admin privileges required"})
			return
		}

		ctx.Next()
	}
}

// SelfOrSuperAdminAuthMiddleware is a middleware that ensures the admin is either modifying their own account
// or is a super admin
func SelfOrSuperAdminAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get admin ID from request
		requestedID := ctx.Param("id")

		// Get admin from context
		adminInterface, exists := ctx.Get("admin")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Get user ID from context
		userID, exists := ctx.Get("userID")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Check if admin is modifying their own account or is a super admin
		admin, ok := adminInterface.(*models.AdminResponse)
		if !ok || (!admin.IsSuper && userID.(string) != requestedID) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Not authorized to modify this admin"})
			return
		}

		ctx.Next()
	}
}
