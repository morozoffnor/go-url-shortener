package app

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
	urlLib "net/url"
	"strings"
)

var urlStorage = URLStorage{
	list: make(map[string]string),
}

func ShortURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Failed parsing body", http.StatusBadRequest)
		return
	}
	decodedBody, err := urlLib.QueryUnescape(string(body))
	if err != nil {
		http.Error(w, "Failed decoding body", http.StatusBadRequest)
		return
	}
	decodedBody, _ = strings.CutPrefix(decodedBody, "url=")
	url, err := urlStorage.addNewURL(decodedBody)
	if err != nil {
		http.Error(w, "Unexpected internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain, utf-8")
	w.WriteHeader(http.StatusCreated)
	_, err = fmt.Fprint(w, "http://localhost:8080/"+url)
	if err != nil {
		log.Print("error while writing response")
		return
	}
}

func FullURL(w http.ResponseWriter, r *http.Request) {
	v, err := urlStorage.getFullURL(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, v, http.StatusTemporaryRedirect)
}

func RunServer() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", FullURL)
	r.Post("/", ShortURL)

	return http.ListenAndServe(`:8080`, r)
}
