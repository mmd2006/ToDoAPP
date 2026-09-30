package main

import (
	"ToDoAPP/config"
	"ToDoAPP/controller"
	"ToDoAPP/repository"
	"ToDoAPP/router"
	"ToDoAPP/service"

	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env")
	}

	config.ConnectMongo()

	if err := config.CreateTaskIndexes(); err != nil {
		log.Fatal("failed to create task indexes:", err)
	}
	config.ConnectMySQL()

	taskRepo := repository.NewTaskMongoRepository(
		config.TaskCollection,
	)

	taskService := service.NewTaskMongoService(
		taskRepo,
	)

	taskController := controller.NewTaskController(
		taskService,
	)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	router.AuthRoutes(e)
	router.InitRoutes(e, taskController)

	port := os.Getenv("PORT")
	if port == "" {
		port = "1323"
	}

	log.Fatal(e.Start(":" + port))
}
