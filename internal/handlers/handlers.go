package handlers

import (
	"fmt"
	"github.com/morozoffnor/go-url-shortener/internal/storage"
	"io"
	"log"
	"net/http"
	urlLib "net/url"
	"strings"
)

var ResponseAddr string

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
	url, err := storage.URLs.AddNewURL(decodedBody)
	if err != nil {
		http.Error(w, "Unexpected internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain, utf-8")
	w.WriteHeader(http.StatusCreated)
	_, err = fmt.Fprint(w, ResponseAddr+"/"+url)
	if err != nil {
		log.Print("error while writing response")
		return
	}
}

func FullURL(w http.ResponseWriter, r *http.Request) {
	v, err := storage.URLs.GetFullURL(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", v)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
