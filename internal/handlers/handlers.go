package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/storage"
	"github.com/morozoffnor/go-url-shortener/pkg/body"
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
	body, err := body.GetBody(r)
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
		var pgErr *pgconn.PgError
		// возвращаем 409 если такой URL уже есть в бд
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			w.Header().Set("Content-Type", "text/plain, utf-8")
			w.WriteHeader(http.StatusConflict)
			// просто Fprint подставляет /n в конце строки, автотесты ругаются
			_, err = fmt.Fprintf(w, "%s", h.cfg.ResultAddr+"/"+url)
			if err != nil {
				log.Print("error while writing response")
				return
			}
			return
		}
		http.Error(w, "Unexpected internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain, utf-8")
	w.WriteHeader(http.StatusCreated)
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
}

func (h *Handlers) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		URL string `json:"url"`
	}
	type resBody struct {
		Result string `json:"result"`
	}
	w.Header().Set("Content-Type", "application/json")
	var raw bytes.Buffer
	if _, err := raw.ReadFrom(r.Body); err != nil {
		http.Error(w, "Invalid body", http.StatusUnprocessableEntity)
		return
	}

	rbody := &reqBody{}
	err := json.Unmarshal(raw.Bytes(), rbody)
	if err != nil {
		http.Error(w, "Invalid json", http.StatusUnprocessableEntity)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	url, err := h.store.AddNewURL(ctx, rbody.URL)

	if err != nil {
		var pgErr *pgconn.PgError
		// возвращаем 409 если такой URL уже есть в бд
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			short := &resBody{Result: h.cfg.ResultAddr + "/" + url}
			resp, err := json.Marshal(short)
			if err != nil {
				http.Error(w, "Fail during serializing", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusConflict)
			w.Write(resp)
			return
		}
		http.Error(w, "Unexpected internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	short := &resBody{Result: h.cfg.ResultAddr + "/" + url}
	resp, err := json.Marshal(short)
	if err != nil {
		http.Error(w, "Fail during serializing", http.StatusInternalServerError)
		return
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

func (h *Handlers) BatchHandler(w http.ResponseWriter, r *http.Request) {
	body, err := body.GetBody(r)
	if err != nil {
		http.Error(w, "Failed parsing body", http.StatusBadRequest)
		return
	}
	var input []storage.BatchInput
	err = json.Unmarshal(body, &input)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		http.Error(w, "Failed decoding body", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	output, err := h.store.AddBatch(ctx, input)
	if err != nil {
		http.Error(w, "Unexpected internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	resp, err := json.Marshal(output)
	if err != nil {
		http.Error(w, "Fail during serializing", http.StatusInternalServerError)
	}
	w.Write(resp)
}
