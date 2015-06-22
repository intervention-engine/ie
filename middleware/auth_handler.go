package middleware

import (
	"fmt"
	"net/http"
)

func AuthHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	auth_token := r.Header.Get("Authorization")

	fmt.Println(auth_token)
	next(rw, r)
}
