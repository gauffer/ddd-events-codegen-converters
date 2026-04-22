package middleware

import "net/http"

const ContentTypeHeaderKey = "Content-Type"

// ContentType — устанавливает заголовок Content-Type
func ContentType(mime string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(ContentTypeHeaderKey, mime)
			next.ServeHTTP(w, r)
		})
	}
}
