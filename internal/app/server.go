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
	//if r.Header.Get("Content-Type") != "text/plain" {
	//	http.Error(w, "Only text/plain is accepted", http.StatusBadRequest)
	//	return
	//}
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
	//w.Header().Write(w)
	_, err = fmt.Fprint(w, "http://localhost:8080/"+url)
	if err != nil {
		log.Print("error while writing response")
		return
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	v, err := urlStorage.getFullUrl(r.PathValue("id"))
	if err != nil {
		log.Print("full url (server error) - " + v)
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}
	log.Print("full url (server) - " + v)
	w.Header().Set("Location", v)
	w.WriteHeader(307)
	log.Print("Location - " + w.Header().Get("Location"))

}

func RunServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", shortURL)
	mux.HandleFunc("GET /{id}/", redirect)
	return http.ListenAndServe(`:8080`, mux)
}
