package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/morozoffnor/go-url-shortener/internal/config"
	"net/http"
	"time"
)

type ContextUserIDKey string

var ContextUserID = "user_id"

type JWT struct {
	config *config.Config
}

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

func New(config *config.Config) *JWT {
	return &JWT{config: config}
}

func (h *JWT) GenerateToken() (string, error) {
	userID := uuid.New()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Hour)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		return "", err
	}

	return "Bearer " + tokenString, nil
}

func (h *JWT) ParseToken(token string) (*Claims, error) {
	token = token[7:]
	claims := &Claims{}
	t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (h *JWT) CheckToken(r *http.Request) bool {
	cookie, err := r.Cookie("Authorization")

	if err != nil {
		return false
	}

	claims, err := h.ParseToken(cookie.Value)
	if err != nil {
		return false
	}
	if claims.UserID == uuid.Nil {
		return false
	}
	return true
}

func (h *JWT) AddTokenToCookies(w *http.ResponseWriter, r *http.Request) (context.Context, error) {
	token, err := h.GenerateToken()

	if err != nil {
		return nil, err
	}

	http.SetCookie(*w, &http.Cookie{
		Name:    "Authorization",
		Value:   token,
		Expires: time.Now().Add(5 * time.Hour),
	})

	claims, err := h.ParseToken(token)
	if err != nil {
		return nil, err
	}

	ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)

	return ctx, nil
}
