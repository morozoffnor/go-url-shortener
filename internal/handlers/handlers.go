package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/storage"
	"github.com/morozoffnor/go-url-shortener/pkg"
	"log"
	"net/http"
	urlLib "net/url"
	"strings"
	"time"
)

type Handlers struct {
	cfg   *config.Config
	store storage.Storage
}

func New(cfg *config.Config, store storage.Storage) *Handlers {
	h := &Handlers{
		cfg:   cfg,
		store: store,
	}

	return h
}

func (h *Handlers) ShortURLHandler(w http.ResponseWriter, r *http.Request) {
	body, err := pkg.GetBody(r)
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

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	url, err := h.store.AddNewURL(ctx, decodedBody)
	defer cancel()

	if err != nil {
		http.Error(w, "Unexpected internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain, utf-8")
	w.WriteHeader(http.StatusCreated)
	//log.Print(storage.URLs.List)
	_, err = fmt.Fprint(w, h.cfg.ResultAddr+"/"+url)
	if err != nil {
		log.Print("error while writing response")
		return
	}
}

func (h *Handlers) FullURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	v, err := h.store.GetFullURL(ctx, r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, v, http.StatusTemporaryRedirect)
	log.Print(w.Header().Get("location"))
}

func (h *Handlers) ShortenHandler(w http.ResponseWriter, r *http.Request) {
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
	//log.Print(body)
	if err != nil {
		http.Error(w, "Invalid json", http.StatusUnprocessableEntity)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	url, err := h.store.AddNewURL(ctx, body.URL)
	//log.Print(storage.URLs.List)
	if err != nil {
		http.Error(w, "Unexpected internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	short := &resBody{Result: h.cfg.ResultAddr + "/" + url}
	resp, err := json.Marshal(short)
	if err != nil {
		http.Error(w, "Fail during serializing", http.StatusInternalServerError)
	}
	w.Write(resp)
}

func (h *Handlers) PingHandler(w http.ResponseWriter, r *http.Request) {
	if h.store.Ping(r.Context()) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
