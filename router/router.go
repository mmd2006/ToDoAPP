package router

import (
	"ToDoAPP/controller"
	"ToDoAPP/middleware"
	"net/http"

	"github.com/labstack/echo/v4"
)

func InitRoutes(
	e *echo.Echo,
	taskController *controller.TaskController,
) {

	e.GET("/", func(c echo.Context) error {
		return c.JSON(
			http.StatusOK,
			map[string]string{
				"message": "ToDoAPP API is running",
			},
		)
	})

	e.GET(
		"/admin/login-activities",
		controller.GetLoginActivities,
		middleware.JWTMiddleware,
		middleware.RequireRole("admin"),
	)

	adminGroup := e.Group(
		"/admin",
		middleware.JWTMiddleware,
	)

	adminGroup.GET(
		"/tasks",
		taskController.GetTasks,
		middleware.RequireRole("admin"),
	)

	adminGroup.GET(
		"/users",
		controller.GetUsers,
		middleware.RequireRole("admin"),
	)

	userGroup := e.Group(
		"/tasks",
		middleware.JWTMiddleware,
	)

	userGroup.POST(
		"",
		taskController.CreateTask,
	)

	userGroup.GET(
		"",
		taskController.GetTasks,
	)

	userGroup.GET(
		"/:id",
		taskController.GetTaskByID,
	)

	userGroup.PUT(
		"/:id",
		taskController.UpdateTask,
	)

	userGroup.DELETE(
		"/:id",
		taskController.DeleteTask,
	)
}

func AuthRoutes(e *echo.Echo) {
	e.POST("/signup", controller.Signup)
	e.POST("/login", controller.Login)
}
