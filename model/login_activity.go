package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LoginActivity struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    uint               `bson:"user_id"`
	Username  string             `bson:"username"`
	IP        string             `bson:"ip"`
	UserAgent string             `bson:"user_agent"`
	CreatedAt time.Time          `bson:"created_at"`
}
