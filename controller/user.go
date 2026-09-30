package controller

import (
	"ToDoAPP/config"
	"ToDoAPP/model"
	mysql "ToDoAPP/repository"
	"ToDoAPP/validation"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Signup(c echo.Context) error {
	var input validation.SignupInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	if err := input.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	repo := mysql.NewUserRepository()

	if _, err := repo.FindByUsername(input.Username); err == nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "username already exist",
		})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "database error",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "failed to hash password",
		})
	}
	user := model.UserSQL{
		Username: input.Username,
		Password: string(hashedPassword),
		Role:     model.RoleUser,
	}

	if err := repo.Create(&user); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "database error"})
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "user created"})
}

func Login(c echo.Context) error {
	var input validation.SignupInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	repo := mysql.NewUserRepository()

	user, err := repo.FindByUsername(input.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusUnauthorized, echo.Map{
				"error": "invalid credentials",
			})
		}

		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "database error",
		})
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(input.Password),
	) != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid credentials"})
	}

	ctx, cancel := context.WithTimeout(
		c.Request().Context(),
		5*time.Second,
	)
	defer cancel()

	activity := model.LoginActivity{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		Username:  input.Username,
		IP:        c.RealIP(),
		UserAgent: c.Request().UserAgent(),
		CreatedAt: time.Now(),
	}

	if _, err := config.LoginActivityCollection.InsertOne(ctx, activity); err != nil {
		log.Printf("failed to record login activity: %v", err)
	}

	claims := jwt.MapClaims{
		"user_id":  fmt.Sprintf("%d", user.ID),
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "JWT secret not configured",
		})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "failed to sign JWT",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": signedToken})
}

func GetUsers(c echo.Context) error {
	repo := mysql.NewUserRepository()

	users, err := repo.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "database error",
		})
	}

	response := make([]map[string]interface{}, len(users))
	for i, u := range users {
		response[i] = map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"role":     u.Role,
		}
	}

	return c.JSON(http.StatusOK, echo.Map{"users": response})
}
