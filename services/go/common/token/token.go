package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Claims struct {
	UserID       string `json:"user_id"`
	InmateNumber string `json:"inmate_number"`
	IssuedAt     int64  `json:"iat"`
	ExpiresAt    int64  `json:"exp"`
}

func Generate(secret []byte, userID, inmateNumber string, ttl time.Duration) (string, int64, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)
	claims := Claims{
		UserID:       userID,
		InmateNumber: inmateNumber,
		IssuedAt:     now.Unix(),
		ExpiresAt:    expiresAt.Unix(),
	}
	rawPayload, err := json.Marshal(claims)
	if err != nil {
		return "", 0, err
	}

	payload := base64.RawURLEncoding.EncodeToString(rawPayload)
	sig := sign(secret, payload)
	token := payload + "." + sig
	return token, expiresAt.Unix(), nil
}

func Parse(secret []byte, raw string) (*Claims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid token format")
	}
	payload, sig := parts[0], parts[1]
	if !hmac.Equal([]byte(sig), []byte(sign(secret, payload))) {
		return nil, errors.New("invalid token signature")
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	var claims Claims
	if err := json.Unmarshal(rawPayload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return nil, errors.New("token expired")
	}
	return &claims, nil
}

func sign(secret []byte, payload string) string {
	h := hmac.New(sha256.New, secret)
	_, _ = h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
