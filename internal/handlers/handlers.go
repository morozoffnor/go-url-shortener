package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	authHelper "github.com/morozoffnor/go-url-shortener/internal/auth"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"github.com/morozoffnor/go-url-shortener/internal/storage"
	"github.com/morozoffnor/go-url-shortener/pkg/body"
	"github.com/morozoffnor/go-url-shortener/pkg/logger"
	"log"
	"net/http"
	urlLib "net/url"
	"time"
)

type Handlers struct {
	Cfg   *config.Config
	store storage.Storage
	auth  *authHelper.JWT
}

func New(cfg *config.Config, store storage.Storage, authHelper *authHelper.JWT) *Handlers {
	h := &Handlers{
		Cfg:   cfg,
		store: store,
		auth:  authHelper,
	}

	return h
}

func (h *Handlers) ShortURLHandler(w http.ResponseWriter, r *http.Request) {
	if !h.auth.CheckToken(r) {
		ctx, err := h.auth.AddTokenToCookies(&w, r)
		if err != nil {
			http.Error(w, "Error creating token", http.StatusBadRequest)
			return
		}
		r = r.WithContext(ctx)
	}
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
			_, err = fmt.Fprintf(w, "%s", h.Cfg.ResultAddr+"/"+url)
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
	_, err = fmt.Fprint(w, h.Cfg.ResultAddr+"/"+url)
	if err != nil {
		log.Print("error while writing response")
		return
	}
}

func (h *Handlers) FullURLHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	v, isDeleted, err := h.store.GetFullURL(ctx, r.PathValue("id"))
	if err != nil {
		logger.Logger.Error(err)
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}
	if isDeleted {
		http.Error(w, "Deleted", http.StatusGone)
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

	if !h.auth.CheckToken(r) {
		ctx, err := h.auth.AddTokenToCookies(&w, r)
		if err != nil {
			http.Error(w, "Error creating token", http.StatusBadRequest)
			return
		}
		r = r.WithContext(ctx)
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
			short := &resBody{Result: h.Cfg.ResultAddr + "/" + url}
			resp, err := json.Marshal(short)
			if err != nil {
				logger.Logger.Error(err)
				http.Error(w, "Fail during serializing", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusConflict)
			w.Write(resp)
			return
		}
		logger.Logger.Error(err)
		http.Error(w, "Unexpected internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	short := &resBody{Result: h.Cfg.ResultAddr + "/" + url}
	resp, err := json.Marshal(short)
	if err != nil {
		logger.Logger.Error(err)
		http.Error(w, "Fail during serializing", http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *Handlers) PingHandler(w http.ResponseWriter, r *http.Request) {
	v, ok := interface{}(h.store).(storage.Pingable)
	if ok {
		if v.Ping(r.Context()) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}

}

func (h *Handlers) BatchHandler(w http.ResponseWriter, r *http.Request) {
	if !h.auth.CheckToken(r) {
		ctx, err := h.auth.AddTokenToCookies(&w, r)
		if err != nil {
			http.Error(w, "Error creating token", http.StatusBadRequest)
			return
		}
		r = r.WithContext(ctx)
	}
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

func (h *Handlers) GetUserURLsHandler(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(authHelper.ContextUserID).(uuid.UUID)

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	v, err := h.store.GetUserURLs(ctx, userID)
	if err != nil {
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if v == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	resp, err := json.Marshal(v)
	log.Print(v)
	if err != nil {
		http.Error(w, "Fail during serializing", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handlers) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	if !h.auth.CheckToken(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var ids []string
	//var items []*storage.DeleteURLItem

	rb, err := body.GetBody(r)
	if err != nil {
		http.Error(w, "Error parsing body", http.StatusBadRequest)
	}
	err = json.Unmarshal(rb, &ids)
	if err != nil {
		http.Error(w, "Error parsing body", http.StatusBadRequest)
	}

	if len(ids) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	h.store.DeleteURLs(r.Context(), r.Context().Value(authHelper.ContextUserID).(uuid.UUID), ids)
	w.WriteHeader(http.StatusAccepted)
}
