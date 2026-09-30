package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "missing token"})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid authorization header format"})
		}
		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "missing token"})
		}

		claims := jwt.MapClaims{}
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "JWT secret not configured"})
		}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid or expired token"})
		}
		if _, ok := claims["user_id"].(string); !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"message": "invalid token claims",
			})
		}

		if _, ok := claims["role"].(string); !ok {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"message": "invalid token claims",
			})
		}

		c.Set("user", claims)

		return next(c)
	}
}

func RequireRole(requireRole string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get("user")
			claims, ok := user.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": "invalid token claims"})
			}

			role, ok := claims["role"].(string)
			if !ok {
				return c.JSON(http.StatusForbidden, echo.Map{"message": "invalid role format"})
			}
			if role != requireRole {
				return c.JSON(http.StatusForbidden, echo.Map{"message": "you are not allowed to access this resource"})
			}

			return next(c)
		}
	}
}
