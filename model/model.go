package model

import (
	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is...
type User struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	Firstname   string             `bson:"firstName" json:"firstName"`
	Lastname    string             `bson:"lastName" json:"lastName"`
	Email       string             `bson:"email" json:"email"`
	IDNumber    string             `bson:"idNumber" json:"idNumber"`
	PhoneNumber string             `bson:"phoneNumber" json:"phoneNumber"`
	Password    string             `bson:"password" json:"password"`
	Token       string             `bson:"token" json:"token"`
	Exp         int                `bson:"exp" json:"exp"`
	FCMToken    string             `bson:"fcmtoken" json:"fcmtoken"`
}

// Vehicle is...
type Vehicle struct {
	RegistrationNumber string `bson:"registrationNumber" json:"registrationNumber"`
	VehicleClass       string `bson:"vehicleClass" json:"vehicleClass"`
	VehicleModel       string `bson:"vehicleModel" json:"vehicleModel"`
	UserID             string `bson:"userId" json:"userId"`
}

// Payment is...
type Payment struct {
	Amount int `bson:"amount" json:"amount"`
}

// ResponseResult is...
type ResponseResult struct {
	Error  string `json:"error"`
	Result string `json:"result"`
}
