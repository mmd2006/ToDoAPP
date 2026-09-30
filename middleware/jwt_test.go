package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const testSecret = "test_secret_key"

func makeToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	if claims["exp"] == nil {
		claims["exp"] = time.Now().Add(time.Hour).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

func setupRequest(t *testing.T, authHeader string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func okHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"message": "ok"})
}

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", testSecret)
	os.Exit(m.Run())
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	c, rec := setupRequest(t, "")
	if err := JWTMiddleware(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_BadHeaderFormat(t *testing.T) {
	c, rec := setupRequest(t, "Token abc123")
	if err := JWTMiddleware(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	c, rec := setupRequest(t, "Bearer not.a.real.token")
	if err := JWTMiddleware(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	token := makeToken(t, jwt.MapClaims{
		"user_id": "1", "role": "user",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	c, rec := setupRequest(t, "Bearer "+token)
	if err := JWTMiddleware(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_MissingUserIDClaim(t *testing.T) {
	token := makeToken(t, jwt.MapClaims{
		"role": "user",
	})

	c, rec := setupRequest(t, "Bearer "+token)

	if err := JWTMiddleware(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_MissingRoleClaim(t *testing.T) {
	token := makeToken(t, jwt.MapClaims{
		"user_id": "1",
	})

	c, rec := setupRequest(t, "Bearer "+token)

	if err := JWTMiddleware(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	token := makeToken(t, jwt.MapClaims{"user_id": "1", "role": "admin"})
	c, rec := setupRequest(t, "Bearer "+token)
	if err := JWTMiddleware(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	claims, ok := c.Get("user").(jwt.MapClaims)
	if !ok {
		t.Fatal("claims not set in context")
	}
	if claims["role"] != "admin" {
		t.Fatalf("expected role admin, got %v", claims["role"])
	}
}

func TestJWTMiddleware_LowercaseBearer(t *testing.T) {
	token := makeToken(t, jwt.MapClaims{"user_id": "1", "role": "user"})
	c, rec := setupRequest(t, "bearer "+token)
	if err := JWTMiddleware(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", jwt.MapClaims{"role": "admin"})

	if err := RequireRole("admin")(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", jwt.MapClaims{"role": "user"})

	if err := RequireRole("admin")(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireRole_MissingRoleClaim(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", jwt.MapClaims{"user_id": "1"})

	if err := RequireRole("admin")(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestJWTMiddleware_NumericUserIDClaim(t *testing.T) {
	token := makeToken(t, jwt.MapClaims{
		"user_id": 1,
		"role":    "user",
	})

	c, rec := setupRequest(t, "Bearer "+token)

	if err := JWTMiddleware(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTMiddleware_BooleanUserIDClaim(t *testing.T) {
	token := makeToken(t, jwt.MapClaims{
		"user_id": true,
		"role":    "user",
	})

	c, rec := setupRequest(t, "Bearer "+token)

	if err := JWTMiddleware(okHandler)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
