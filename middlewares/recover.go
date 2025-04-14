package middlewares

import (
	"fmt"
	"net/http"
)
func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				fmt.Println(err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}