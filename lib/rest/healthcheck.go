package rest

import "net/http"

func BuildHealthcheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := m{"status": "ok"}
		respond(w, http.StatusOK, resp)
	}
}
