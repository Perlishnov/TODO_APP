package dao

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserDAOMongo struct{
	collection *mongo.Collection
}



func NewUser(db *mongo.Database, logger *logrus.Logger) AuthDAO  {
	collection := db.Collection("users")
	Initiliaze 

	_, err := collection.
}