package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	Id        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	Email     string             `json:"email" bson:"email"`
	Password  string             `json:"password" bson:"password"`
	Admin     bool               `json:"admin" bson:"admin"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
