package controller

import (
	"ToDoAPP/model"
	"ToDoAPP/service"
	"ToDoAPP/validation"
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskController struct {
	service service.TaskMongoService
}

func NewTaskController(taskService service.TaskMongoService) *TaskController {
	return &TaskController{
		service: taskService,
	}
}

func userObjectID(claims jwt.MapClaims) (primitive.ObjectID, error) {
	raw, ok := claims["user_id"]
	if !ok {
		return primitive.NilObjectID, errors.New("missing user_id claim")
	}

	switch v := raw.(type) {
	case string:
		if oid, err := primitive.ObjectIDFromHex(v); err == nil {
			return oid, nil
		}
		return primitive.ObjectIDFromHex(padNumericID(v))
	case float64:
		return primitive.ObjectIDFromHex(padNumericID(formatUint(uint64(v))))
	default:
		return primitive.NilObjectID, errors.New("unsupported user_id type")
	}
}

func formatUint(n uint64) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}

// padNumericID converts a numeric id string into a 24-char hex string that
// is a valid ObjectID (right-aligned, zero-padded).
func padNumericID(s string) string {
	hexDigits := make([]byte, 0, len(s)*2)
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return s // not numeric; let ObjectIDFromHex report the error
		}
		hexDigits = append(hexDigits, '0', byte(ch))
	}
	if len(hexDigits) > 24 {
		hexDigits = hexDigits[len(hexDigits)-24:]
	}
	for len(hexDigits) < 24 {
		hexDigits = append([]byte{'0'}, hexDigits...)
	}
	return string(hexDigits)
}

func (tc *TaskController) CreateTask(c echo.Context) error {
	var input validation.TaskInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}
	if err := input.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	user := c.Get("user")
	claims, ok := user.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
	}

	userID, err := userObjectID(claims)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user ID"})
	}

	task := model.Task{
		ID:          primitive.NewObjectID(),
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		Completed:   false,
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	err = tc.service.CreateTask(ctx, &task)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "database error",
		})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"id":          task.ID.Hex(),
		"user_id":     task.UserID.Hex(),
		"title":       task.Title,
		"description": task.Description,
		"completed":   task.Completed,
	})
}

func (tc *TaskController) GetTasks(c echo.Context) error {
	ctx, cancel := context.WithTimeout(
		c.Request().Context(),
		10*time.Second,
	)
	defer cancel()

	user, ok := c.Get("user").(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error": "invalid token claims",
		})
	}

	role, ok := user["role"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid role in token",
		})
	}

	userID, err := userObjectID(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid user ID",
		})
	}

	page := int64(1)

	if pageParam := c.QueryParam("page"); pageParam != "" {
		parsedPage, err := strconv.ParseInt(pageParam, 10, 64)
		if err != nil || parsedPage < 1 {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid page",
			})
		}

		page = parsedPage
	}

	limit := int64(10)

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		parsedLimit, err := strconv.ParseInt(limitParam, 10, 64)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid limit",
			})
		}

		limit = parsedLimit
	}

	skip := (page - 1) * limit

	tasks, err := tc.service.GetTasks(
		ctx,
		userID,
		role,
		skip,
		limit,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "failed to get tasks",
		})
	}

	response := make([]map[string]interface{}, len(tasks))

	for i, t := range tasks {
		response[i] = map[string]interface{}{
			"id":          t.ID.Hex(),
			"user_id":     t.UserID.Hex(),
			"title":       t.Title,
			"description": t.Description,
			"completed":   t.Completed,
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (tc *TaskController) GetTaskByID(c echo.Context) error {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid task id"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	user, ok := c.Get("user").(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
	}
	role, ok := user["role"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid role in token",
		})
	}

	userID, err := userObjectID(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid user ID",
		})
	}

	task, err := tc.service.GetTaskByID(
		ctx,
		objectID,
		userID,
		role,
	)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.JSON(http.StatusNotFound, echo.Map{
				"error": "task not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "failed to get task",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"id":          task.ID.Hex(),
		"user_id":     task.UserID.Hex(),
		"title":       task.Title,
		"description": task.Description,
		"completed":   task.Completed,
	})
}

func (tc *TaskController) UpdateTask(c echo.Context) error {
	idParam := c.Param("id")

	taskID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid ID format",
		})
	}

	var input validation.TaskInput

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid input",
		})
	}

	if err := input.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}

	user, ok := c.Get("user").(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error": "invalid token claims",
		})
	}

	role, ok := user["role"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid role in token",
		})
	}

	userID, err := userObjectID(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid user ID",
		})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(),
		10*time.Second,
	)
	defer cancel()

	err = tc.service.UpdateTask(
		ctx,
		taskID,
		userID,
		role,
		input.Title,
		input.Description,
	)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.JSON(http.StatusNotFound, echo.Map{
				"error": "task not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "failed to update task",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "task updated",
	})

}

func (tc *TaskController) DeleteTask(c echo.Context) error {
	idParam := c.Param("id")

	taskID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid ID format",
		})
	}

	user, ok := c.Get("user").(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"error": "invalid token claims",
		})
	}

	role, ok := user["role"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid role in token",
		})
	}

	userID, err := userObjectID(user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid user ID",
		})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(),
		10*time.Second,
	)
	defer cancel()

	err = tc.service.DeleteTask(ctx, taskID, userID, role)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.JSON(http.StatusNotFound, echo.Map{
				"error": "task not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "failed to delete task",
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "task deleted",
	})
}
