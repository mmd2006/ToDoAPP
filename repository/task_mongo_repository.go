package repository

import (
	"context"

	"ToDoAPP/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskMongoRepository interface {
	Create(ctx context.Context, task *model.Task) error
	FindAll(
		ctx context.Context,
		filter bson.M,
		skip int64,
		limit int64,
	) ([]model.Task, error)
	FindByID(ctx context.Context, filter bson.M) (*model.Task, error)
	Update(ctx context.Context, filter bson.M, update bson.M) error
	Delete(ctx context.Context, filter bson.M) error
}

type taskMongoRepository struct {
	collection *mongo.Collection
}

func NewTaskMongoRepository(collection *mongo.Collection) TaskMongoRepository {
	return &taskMongoRepository{
		collection: collection,
	}
}

func (r *taskMongoRepository) Create(
	ctx context.Context,
	task *model.Task,
) error {
	_, err := r.collection.InsertOne(ctx, task)
	return err
}

func (r *taskMongoRepository) FindAll(
	ctx context.Context,
	filter bson.M,
	skip int64,
	limit int64,
) ([]model.Task, error) {

	cursor, err := r.collection.Find(
		ctx,
		filter,
		options.Find().
			SetSkip(skip).
			SetLimit(limit).
			SetSort(bson.D{
				{Key: "created_at", Value: -1},
			}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []model.Task

	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *taskMongoRepository) FindByID(
	ctx context.Context,
	filter bson.M,
) (*model.Task, error) {

	var task model.Task

	err := r.collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *taskMongoRepository) Update(
	ctx context.Context,
	filter bson.M,
	update bson.M,
) error {

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *taskMongoRepository) Delete(
	ctx context.Context,
	filter bson.M,
) error {

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

var _ TaskMongoRepository = (*taskMongoRepository)(nil)
