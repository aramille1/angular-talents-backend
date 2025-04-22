package middlewares

import (
	"context"
	"net/http"
	"reverse-job-board/domain"
	"reverse-job-board/internal"
	"strings"
)

// ValidateAdminAuth is a middleware that validates admin JWT tokens
func ValidateAdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		if authorization == "" {
			err := internal.NewError(http.StatusUnauthorized,
				"admin.authentication.retrieve_token",
				"Authentication required",
				"empty authorization header")
			internal.WriteError(w, err)
			return
		}

		// Check if the Authorization header has the Bearer prefix
		if !strings.HasPrefix(authorization, "Bearer ") {
			err := internal.NewError(http.StatusUnauthorized,
				"admin.authentication.invalid_format",
				"Invalid authorization format",
				"authorization header doesn't start with 'Bearer '")
			internal.WriteError(w, err)
			return
		}

		// Extract the token
		splitToken := strings.Split(authorization, "Bearer ")
		authToken := splitToken[1]

		if authToken == "" {
			err := internal.NewError(http.StatusUnauthorized,
				"admin.authentication.empty_token",
				"Authorization token is empty",
				"empty token after Bearer prefix")
			internal.WriteError(w, err)
			return
		}

		// Validate the admin token - we get adminID and role from token
		adminID, _, err := domain.ValidateAdminToken(authToken)
		if err != nil {
			err := internal.NewError(http.StatusUnauthorized,
				"admin.authentication.invalid_token",
				"Invalid authorization token",
				err.Error())
			internal.WriteError(w, err)
			return
		}

		// Since the token was generated specifically for an admin user,
		// and validated successfully, we can add the admin ID to the context
		ctx := context.WithValue(r.Context(), "adminID", adminID)

		// Continue with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
