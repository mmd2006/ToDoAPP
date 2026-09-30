package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
)

var Client *mongo.Client
var DBName string
var UserCollection *mongo.Collection
var TaskCollection *mongo.Collection
var LoginActivityCollection *mongo.Collection

func ConnectMongo() {
	_ = godotenv.Load() // main already loads .env; ignore error here

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI is not set")
	}
	DBName = os.Getenv("MONGO_DB")
	if DBName == "" {
		DBName = "todoapp"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("MongoDB connection error", err)
	}

	log.Println("connected to MongoDB")

	Client = client

	UserCollection = Client.Database(DBName).Collection("users")
	TaskCollection = Client.Database(DBName).Collection("tasks")
	LoginActivityCollection = client.Database(DBName).Collection("login_activity")

}

func CreateTaskIndexes() error {
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	}

	_, err := TaskCollection.Indexes().CreateOne(
		context.Background(),
		indexModel,
	)

	return err
}
