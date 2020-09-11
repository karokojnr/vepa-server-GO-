package util

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"reflect"
	"time"
)

func GetCollection(collectionName string) (*mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(GoDotEnvVariable("MONGO_URI")))
	if err != nil {
		return nil, err
	}
	//log.Println("---Client Type---")
	//log.Println(reflect.TypeOf(client))

	//var collection *mongo.Collection
	//c *mongo.Database
	//db := client.Database("vepadb")
	collection := client.Database("vepadb").Collection(collectionName)
	return collection, nil

}
