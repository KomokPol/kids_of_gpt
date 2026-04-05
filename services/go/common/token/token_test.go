package token

import (
	"testing"
	"time"
)

func TestGenerateAndParse(t *testing.T) {
	secret := []byte("test-secret")
	raw, _, err := Generate(secret, "u-1", "in-1", 5*time.Minute)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	claims, err := Parse(secret, raw)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.UserID != "u-1" {
		t.Fatalf("unexpected user id: %s", claims.UserID)
	}
	if claims.InmateNumber != "in-1" {
		t.Fatalf("unexpected inmate number: %s", claims.InmateNumber)
	}
}
