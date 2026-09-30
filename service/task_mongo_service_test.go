package service

import (
	"context"
	"errors"
	"testing"

	"ToDoAPP/model"
	"ToDoAPP/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mockTaskMongoRepository struct {
	createFunc   func(ctx context.Context, task *model.Task) error
	findAllFunc  func(ctx context.Context, filter bson.M, skip int64, limit int64) ([]model.Task, error)
	findByIDFunc func(ctx context.Context, filter bson.M) (*model.Task, error)
	updateFunc   func(ctx context.Context, filter bson.M, update bson.M) error
	deleteFunc   func(ctx context.Context, filter bson.M) error
}

func (m *mockTaskMongoRepository) Create(ctx context.Context, task *model.Task) error {
	return m.createFunc(ctx, task)
}

func (m *mockTaskMongoRepository) FindAll(
	ctx context.Context,
	filter bson.M,
	skip int64,
	limit int64,
) ([]model.Task, error) {
	return m.findAllFunc(ctx, filter, skip, limit)
}

func (m *mockTaskMongoRepository) FindByID(
	ctx context.Context,
	filter bson.M,
) (*model.Task, error) {
	return m.findByIDFunc(ctx, filter)
}

func (m *mockTaskMongoRepository) Update(
	ctx context.Context,
	filter bson.M,
	update bson.M,
) error {
	return m.updateFunc(ctx, filter, update)
}

func (m *mockTaskMongoRepository) Delete(
	ctx context.Context,
	filter bson.M,
) error {
	return m.deleteFunc(ctx, filter)
}

var _ repository.TaskMongoRepository = (*mockTaskMongoRepository)(nil)

func TestTaskMongoService_GetTasks_User(t *testing.T) {
	userID := primitive.NewObjectID()

	mockRepo := &mockTaskMongoRepository{
		findAllFunc: func(
			ctx context.Context,
			filter bson.M,
			skip int64,
			limit int64,
		) ([]model.Task, error) {

			expectedFilter := bson.M{
				"user_id": userID,
			}

			if len(filter) != len(expectedFilter) {
				t.Fatalf("unexpected filter: %v", filter)
			}

			if filter["user_id"] != userID {
				t.Fatalf("expected user_id %v, got %v", userID, filter["user_id"])
			}

			return []model.Task{}, nil
		},
	}

	svc := NewTaskMongoService(mockRepo)

	_, err := svc.GetTasks(
		context.Background(),
		userID,
		"user",
		0,
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskMongoService_GetTasks_Admin(t *testing.T) {
	userID := primitive.NewObjectID()

	mockRepo := &mockTaskMongoRepository{
		findAllFunc: func(
			ctx context.Context,
			filter bson.M,
			skip int64,
			limit int64,
		) ([]model.Task, error) {

			if len(filter) != 0 {
				t.Fatalf("expected empty filter for admin, got %v", filter)
			}

			return []model.Task{}, nil
		},
	}

	svc := NewTaskMongoService(mockRepo)

	_, err := svc.GetTasks(
		context.Background(),
		userID,
		"admin",
		0,
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskMongoService_UpdateTask_User(t *testing.T) {
	userID := primitive.NewObjectID()
	taskID := primitive.NewObjectID()

	mockRepo := &mockTaskMongoRepository{
		updateFunc: func(
			ctx context.Context,
			filter bson.M,
			update bson.M,
		) error {
			if filter["_id"] != taskID {
				t.Fatalf("expected task_id %v, got %v", taskID, filter["_id"])
			}

			if filter["user_id"] != userID {
				t.Fatalf("expected user_id %v, got %v", userID, filter["user_id"])
			}

			if len(filter) != 2 {
				t.Fatalf("expected 2 filter fields, got %v", filter)
			}

			return nil
		},
	}

	svc := NewTaskMongoService(mockRepo)

	err := svc.UpdateTask(
		context.Background(),
		taskID,
		userID,
		"user",
		"Updated title",
		"Updated description",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskMongoService_DeleteTask_User(t *testing.T) {
	userID := primitive.NewObjectID()
	taskID := primitive.NewObjectID()

	mockRepo := &mockTaskMongoRepository{
		deleteFunc: func(
			ctx context.Context,
			filter bson.M,
		) error {
			if filter["_id"] != taskID {
				t.Fatalf("expected task_id %v, got %v", taskID, filter["_id"])
			}

			if filter["user_id"] != userID {
				t.Fatalf("expected user_id %v, got %v", userID, filter["user_id"])
			}

			if len(filter) != 2 {
				t.Fatalf("expected 2 filter fields, got %v", filter)
			}

			return nil
		},
	}

	svc := NewTaskMongoService(mockRepo)

	err := svc.DeleteTask(
		context.Background(),
		taskID,
		userID,
		"user",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskMongoService_UpdateTask_Admin(t *testing.T) {
	userID := primitive.NewObjectID()
	taskID := primitive.NewObjectID()

	mockRepo := &mockTaskMongoRepository{
		updateFunc: func(
			ctx context.Context,
			filter bson.M,
			update bson.M,
		) error {
			if filter["_id"] != taskID {
				t.Fatalf("expected task_id %v, got %v", taskID, filter["_id"])
			}

			if len(filter) != 1 {
				t.Fatalf("expected only _id in admin filter, got %v", filter)
			}

			if _, exists := filter["user_id"]; exists {
				t.Fatalf("admin filter should not contain user_id: %v", filter)
			}

			return nil
		},
	}

	svc := NewTaskMongoService(mockRepo)

	err := svc.UpdateTask(
		context.Background(),
		taskID,
		userID,
		"admin",
		"Updated title",
		"Updated description",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskMongoService_DeleteTask_Admin(t *testing.T) {
	userID := primitive.NewObjectID()
	taskID := primitive.NewObjectID()

	mockRepo := &mockTaskMongoRepository{
		deleteFunc: func(
			ctx context.Context,
			filter bson.M,
		) error {
			if filter["_id"] != taskID {
				t.Fatalf("expected task_id %v, got %v", taskID, filter["_id"])
			}

			if len(filter) != 1 {
				t.Fatalf("expected only _id in admin filter, got %v", filter)
			}

			if _, exists := filter["user_id"]; exists {
				t.Fatalf("admin filter should not contain user_id: %v", filter)
			}

			return nil
		},
	}

	svc := NewTaskMongoService(mockRepo)

	err := svc.DeleteTask(
		context.Background(),
		taskID,
		userID,
		"admin",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskMongoService_GetTaskByID_RepositoryError(t *testing.T) {
	userID := primitive.NewObjectID()
	taskID := primitive.NewObjectID()

	mockRepo := &mockTaskMongoRepository{
		findByIDFunc: func(
			ctx context.Context,
			filter bson.M,
		) (*model.Task, error) {
			return nil, mongo.ErrNoDocuments
		},
	}

	svc := NewTaskMongoService(mockRepo)

	_, err := svc.GetTaskByID(
		context.Background(),
		taskID,
		userID,
		"user",
	)

	if !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("expected mongo.ErrNoDocuments, got %v", err)
	}
}

func TestTaskMongoService_GetTaskByID_User(t *testing.T) {
	userID := primitive.NewObjectID()
	taskID := primitive.NewObjectID()

	mockRepo := &mockTaskMongoRepository{
		findByIDFunc: func(
			ctx context.Context,
			filter bson.M,
		) (*model.Task, error) {
			if filter["_id"] != taskID {
				t.Fatalf("expected task_id %v, got %v", taskID, filter["_id"])
			}

			if filter["user_id"] != userID {
				t.Fatalf("expected user_id %v, got %v", userID, filter["user_id"])
			}

			if len(filter) != 2 {
				t.Fatalf("expected 2 filter fields, got %v", filter)
			}

			return &model.Task{
				ID:     taskID,
				UserID: userID,
			}, nil
		},
	}

	svc := NewTaskMongoService(mockRepo)

	task, err := svc.GetTaskByID(
		context.Background(),
		taskID,
		userID,
		"user",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task == nil {
		t.Fatal("expected task, got nil")
	}
}

func TestTaskMongoService_GetTaskByID_Admin(t *testing.T) {
	userID := primitive.NewObjectID()
	taskID := primitive.NewObjectID()

	mockRepo := &mockTaskMongoRepository{
		findByIDFunc: func(
			ctx context.Context,
			filter bson.M,
		) (*model.Task, error) {
			if filter["_id"] != taskID {
				t.Fatalf("expected task_id %v, got %v", taskID, filter["_id"])
			}

			if len(filter) != 1 {
				t.Fatalf("expected only _id in admin filter, got %v", filter)
			}

			if _, exists := filter["user_id"]; exists {
				t.Fatalf("admin filter should not contain user_id: %v", filter)
			}

			return &model.Task{
				ID:     taskID,
				UserID: userID,
			}, nil
		},
	}

	svc := NewTaskMongoService(mockRepo)

	task, err := svc.GetTaskByID(
		context.Background(),
		taskID,
		userID,
		"admin",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task == nil {
		t.Fatal("expected task, got nil")
	}
}
