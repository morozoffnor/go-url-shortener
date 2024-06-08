package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	url2 "net/url"
	"strings"
)

var urlStorage = UrlStorage{
	list: make(map[string]string),
}

func ShortURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed parsing body", http.StatusBadRequest)
		return
	}
	decodedBody, err := url2.QueryUnescape(string(body))
	if err != nil {
		http.Error(w, "Failed decoding body", http.StatusBadRequest)
		return
	}
	decodedBody, _ = strings.CutPrefix(decodedBody, "url=")
	url, err := urlStorage.addNewUrl(decodedBody)
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

func FullUrl(w http.ResponseWriter, r *http.Request) {
	v, err := urlStorage.getFullUrl(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}
	//w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	//w.Header().Set("Location", v)
	//w.WriteHeader(307)
	http.Redirect(w, r, v, 307)
}

func RunServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", ShortURL)
	mux.HandleFunc("GET /{id}", FullUrl)
	return http.ListenAndServe(`:8080`, mux)
}
