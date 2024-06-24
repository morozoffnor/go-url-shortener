package handlers

import (
	"bytes"
	"encoding/json"
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
	log.Print(storage.URLs.List)
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
	http.Redirect(w, r, v, http.StatusTemporaryRedirect)
	log.Print(w.Header().Get("location"))
}

func Shorten(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		URL string `json:"url"`
	}
	type resBody struct {
		Result string `json:"result"`
	}
	var raw bytes.Buffer
	if _, err := raw.ReadFrom(r.Body); err != nil {
		http.Error(w, "Invalid body", http.StatusUnprocessableEntity)
		return
	}

	body := &reqBody{}
	err := json.Unmarshal(raw.Bytes(), body)
	if err != nil {
		http.Error(w, "Invalid json", http.StatusUnprocessableEntity)
		return
	}

	url, err := storage.URLs.AddNewURL(body.URL)
	log.Print(storage.URLs.List)
	if err != nil {
		http.Error(w, "Unexpected internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	short := &resBody{Result: ResponseAddr + "/" + url}
	resp, err := json.Marshal(short)
	if err != nil {
		http.Error(w, "Fail during serializing", http.StatusInternalServerError)
	}
	w.Write(resp)
}
