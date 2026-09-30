package controller

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestUserObjectID_HexString(t *testing.T) {
	claims := jwt.MapClaims{"user_id": "665f1a2b3c4d5e6f7a8b9c0d"}
	oid, err := userObjectID(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if oid.Hex() != "665f1a2b3c4d5e6f7a8b9c0d" {
		t.Fatalf("unexpected oid: %s", oid.Hex())
	}
}

func TestUserObjectID_NumericString(t *testing.T) {
	claims := jwt.MapClaims{"user_id": "42"}
	oid, err := userObjectID(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if oid.IsZero() {
		t.Fatal("expected non-zero ObjectID")
	}
	// same input must be deterministic
	oid2, _ := userObjectID(claims)
	if oid != oid2 {
		t.Fatal("expected deterministic conversion")
	}
}

func TestUserObjectID_Float(t *testing.T) {
	claims := jwt.MapClaims{"user_id": float64(7)}
	oid, err := userObjectID(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if oid.IsZero() {
		t.Fatal("expected non-zero ObjectID")
	}
}

func TestUserObjectID_Missing(t *testing.T) {
	if _, err := userObjectID(jwt.MapClaims{}); err == nil {
		t.Fatal("expected error for missing user_id")
	}
}

func TestUserObjectID_BadType(t *testing.T) {
	if _, err := userObjectID(jwt.MapClaims{"user_id": true}); err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
