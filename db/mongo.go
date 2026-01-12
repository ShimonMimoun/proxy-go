package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var Collection *mongo.Collection

func Init(uri, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}

	Client = client
	Collection = client.Database(dbName).Collection("logs")
	log.Println("Connected to MongoDB")
}

type LogEntry struct {
	Timestamp      time.Time   `bson:"timestamp"`
	Method         string      `bson:"method"`
	Path           string      `bson:"path"`
	RemoteIP       string      `bson:"remote_ip"`
	RequestHeader  interface{} `bson:"request_header"`
	RequestBody    interface{} `bson:"request_body"`
	ResponseStatus int         `bson:"response_status"`
	ResponseBody   interface{} `bson:"response_body"`
	DurationMs     int64       `bson:"duration_ms"`
	Provider       string      `bson:"provider"` // "azure" or "bedrock"
}

func LogExchange(entry LogEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Collection.InsertOne(ctx, entry)
	if err != nil {
		log.Println("Failed to log to MongoDB:", err)
	}
}
