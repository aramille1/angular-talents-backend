package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"angular-talents-backend/domain"
	"angular-talents-backend/internal"
	"strings"
)

func ValidateAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		if authorization == "" {
			err := internal.NewError(http.StatusBadRequest,"authentication.retrieve_token", "failed to validate authentication", "empty authorization header")
			internal.WriteError(w, err)
			return
		}

		splitToken := strings.Split(authorization, "Bearer ")
		authToken := splitToken[1]

		if authToken == "" {
			fmt.Println("No authorization token")
			err := internal.NewError(http.StatusBadRequest,"authentication.retrieve_token", "failed to validate authentication", "no authorization token")
			internal.WriteError(w, err)
			return
		}

		userID, err := domain.ValidateToken(authToken)
		if err != nil {
			fmt.Println("Invalid token")
			err := internal.NewError(http.StatusBadRequest,"authentication.validate_token", "failed to validate authentication", "invalid authorization token")
			internal.WriteError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
