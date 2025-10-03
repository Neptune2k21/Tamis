package api

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/emails", emailsHandler)
	mux.HandleFunc("/delete", deleteHandler)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login endpoint"))
}

func emailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Emails endpoint"))
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Delete endpoint"))
}
