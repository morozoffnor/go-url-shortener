package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

var urlStorage = UrlStorage{
	list: make(map[string]string),
}

func shortURL(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Only text/plain is accepted", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed parsing body", http.StatusBadRequest)
		return
	}
	url, err := urlStorage.addNewUrl(string(body))
	if err != nil {
		http.Error(w, "Unexpected internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	_, err = fmt.Fprint(w, "http://localhost:8080/"+url)
	if err != nil {
		log.Print("error while writing response")
		return
	}
}

func RunServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", shortURL)
	return http.ListenAndServe(`:8080`, mux)
}
