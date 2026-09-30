package controller

import (
	"ToDoAPP/config"
	"ToDoAPP/model"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetLoginActivities(c echo.Context) error {
	page := 1
	if pageParam := c.QueryParam("page"); pageParam != "" {
		parsedPage, err := strconv.Atoi(pageParam)
		if err != nil || parsedPage < 1 {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid page"})
		}
		page = parsedPage
	}

	limit := 10
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err != nil || parsedLimit < 1 || parsedLimit > 50 {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid limit"})
		}
		limit = parsedLimit
	}

	skip := (page - 1) * limit

	ctx, cancel := context.WithTimeout(
		c.Request().Context(),
		10*time.Second,
	)
	defer cancel()

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := config.LoginActivityCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to read data"})
	}

	var activities []model.LoginActivity
	if err := cursor.All(ctx, &activities); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to read data"})
	}

	response := make([]map[string]interface{}, 0, len(activities))
	for _, a := range activities {
		response = append(response, map[string]interface{}{
			"id":         a.ID.Hex(),
			"user_id":    a.UserID,
			"username":   a.Username,
			"ip":         a.IP,
			"user_agent": a.UserAgent,
			"created_at": a.CreatedAt,
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"page":  page,
		"limit": limit,
		"data":  response,
	})
}
