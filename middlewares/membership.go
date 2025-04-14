package middlewares

import (
	"context"
	"net/http"
	"angular-talents-backend/dao"
	"angular-talents-backend/domain"
	"angular-talents-backend/internal"
	"strings"
)

func ValidateMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		if authorization == "" {
			ctx := context.WithValue(r.Context(), "userID", "")
			ctx = context.WithValue(ctx, "isMember", false)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		splitToken := strings.Split(authorization, "Bearer ")
		authToken := splitToken[1]

		userID, err := domain.ValidateToken(authToken)
		if err != nil {
			ctx := context.WithValue(r.Context(), "userID", "")
			ctx = context.WithValue(ctx, "isMember", false)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		recruiter, err := dao.FindRecruiterByUser(r.Context(), userID)
		if err != nil {
			err :=  internal.NewError(http.StatusInternalServerError, "membership.find_recruiter", "failed to retrieve recruiter", err.Error())
			internal.WriteError(w, err)
			return
		}
	
		if recruiter == nil {
			ctx := context.WithValue(r.Context(), "userID", userID)
			ctx = context.WithValue(ctx, "isMember", false)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "isMember", recruiter.IsMember)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
