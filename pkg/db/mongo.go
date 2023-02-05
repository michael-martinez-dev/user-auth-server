package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoConnection interface {
	Close()
	DB() *mongo.Database
}

type mongoConn struct {
	client  *mongo.Client
	session mongo.Session
}

func NewMongoConnection() MongoConnection {
	var c mongoConn
	var err error
	url := getURL()
	c.client, err = mongo.Connect(context.TODO(),
		options.Client().ApplyURI(url))
	if err != nil {
		log.Panicln(err.Error())
	}
	if err := c.client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	c.session, err = c.client.StartSession()
	if err != nil {
		log.Panicln(err.Error())
	}
	return &c
}

func (c *mongoConn) Close() {
	c.session.EndSession(context.TODO())
}

func (c *mongoConn) DB() *mongo.Database {
	return c.client.Database(os.Getenv("DATABASE_NAME"))
}

func getURL() string {
	port, err := strconv.Atoi(os.Getenv("DATABASE_PORT"))
	if err != nil {
		return fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority",
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_PASS"),
			os.Getenv("DATABASE_HOST"),
		)
	}
	return fmt.Sprintf("mongodb://%s:%s@%s:%d",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASS"),
		os.Getenv("DATABASE_HOST"),
		port,
	)
}
