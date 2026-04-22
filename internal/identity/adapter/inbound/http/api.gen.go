package http

import "net/http"

// ServerInterface — сгенерировано из OpenAPI
type ServerInterface interface {
	PhoneInitiate(w http.ResponseWriter, r *http.Request)
	ResendPhoneCode(w http.ResponseWriter, r *http.Request)
	PhoneVerify(w http.ResponseWriter, r *http.Request)
	Token(w http.ResponseWriter, r *http.Request)
}

// HandlerFromMux — роутинг для ServerInterface
func HandlerFromMux(si ServerInterface, mux *http.ServeMux) http.Handler {
	mux.HandleFunc("POST /v1/identity/oauth/phone/initiate", si.PhoneInitiate)
	mux.HandleFunc("POST /v1/identity/oauth/phone/resend", si.ResendPhoneCode)
	mux.HandleFunc("POST /v1/identity/oauth/phone/verify", si.PhoneVerify)
	mux.HandleFunc("POST /v1/identity/oauth/token", si.Token)
	return mux
}
