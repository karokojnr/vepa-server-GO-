package util

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func GetCollection(collectionName string) (*mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(GoDotEnvVariable("MONGO_URI")))
	if err != nil {
		return nil, err
	}
	Log("MONGO CONNECTED")
	collection := client.Database("vepa").Collection(collectionName)
	return collection, nil

}
