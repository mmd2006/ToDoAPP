package service

import (
	"context"
	"time"

	"ToDoAPP/model"
	"ToDoAPP/repository"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskMongoService interface {
	CreateTask(ctx context.Context, task *model.Task) error
	GetTasks(
		ctx context.Context,
		userID primitive.ObjectID,
		role string,
		skip int64,
		limit int64,
	) ([]model.Task, error)

	GetTaskByID(
		ctx context.Context,
		taskID primitive.ObjectID,
		userID primitive.ObjectID,
		role string,
	) (*model.Task, error)

	UpdateTask(
		ctx context.Context,
		taskID primitive.ObjectID,
		userID primitive.ObjectID,
		role string,
		title string,
		description string,
	) error

	DeleteTask(
		ctx context.Context,
		taskID primitive.ObjectID,
		userID primitive.ObjectID,
		role string,
	) error
}

type taskMongoService struct {
	repo repository.TaskMongoRepository
}

func NewTaskMongoService(
	repo repository.TaskMongoRepository,
) TaskMongoService {
	return &taskMongoService{
		repo: repo,
	}
}

func (s *taskMongoService) CreateTask(
	ctx context.Context,
	task *model.Task,
) error {

	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	return s.repo.Create(ctx, task)
}

func (s *taskMongoService) GetTasks(
	ctx context.Context,
	userID primitive.ObjectID,
	role string,
	skip int64,
	limit int64,
) ([]model.Task, error) {

	var filter bson.M

	if role == model.RoleAdmin {
		filter = bson.M{}
	} else {
		filter = bson.M{
			"user_id": userID,
		}
	}

	return s.repo.FindAll(ctx, filter, skip, limit)
}

func (s *taskMongoService) GetTaskByID(
	ctx context.Context,
	taskID primitive.ObjectID,
	userID primitive.ObjectID,
	role string,
) (*model.Task, error) {

	filter := bson.M{
		"_id": taskID,
	}

	if role != model.RoleAdmin {
		filter["user_id"] = userID
	}

	return s.repo.FindByID(ctx, filter)
}

func (s *taskMongoService) UpdateTask(
	ctx context.Context,
	taskID primitive.ObjectID,
	userID primitive.ObjectID,
	role string,
	title string,
	description string,
) error {

	filter := bson.M{
		"_id": taskID,
	}

	if role != model.RoleAdmin {
		filter["user_id"] = userID
	}

	update := bson.M{
		"$set": bson.M{
			"title":       title,
			"description": description,
		},
	}

	return s.repo.Update(ctx, filter, update)
}

func (s *taskMongoService) DeleteTask(
	ctx context.Context,
	taskID primitive.ObjectID,
	userID primitive.ObjectID,
	role string,
) error {

	filter := bson.M{
		"_id": taskID,
	}

	if role != model.RoleAdmin {
		filter["user_id"] = userID
	}

	return s.repo.Delete(ctx, filter)
}
