package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// GetUserCollection is...
func GetUserCollection() (*mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://karokojnr:karokojnr@cluster0-ubthk.gcp.mongodb.net"))
	if err != nil {
		return nil, err
	}
	collection := client.Database("vepadb").Collection("users")
	return collection, nil
}

// GetVehicleCollection is...
func GetVehicleCollection() (*mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://karokojnr:karokojnr@cluster0-ubthk.gcp.mongodb.net"))
	if err != nil {
		return nil, err
	}
	collection := client.Database("vepadb").Collection("vehicles")
	return collection, nil
}

// GetPaymentCollection is...
func GetPaymentCollection() (*mongo.Collection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://karokojnr:karokojnr@cluster0-ubthk.gcp.mongodb.net"))
	if err != nil {
		return nil, err
	}
	collection := client.Database("vepadb").Collection("payments")
	return collection, nil
}
