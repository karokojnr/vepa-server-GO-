package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// User is...
type User struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	Firstname   string             `json:"firstName"`
	Lastname    string             `json:"lastName"`
	Email       string             `json:"email"`
	IDNumber    string             `json:"idNumber"`
	PhoneNumber string             `json:"phoneNumber"`
	Password    string             `json:"password"`
	Token       string             `json:"token"`
	Exp         int                `json:"exp"`
	FCMToken    string             `json:"fcmtoken"`
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
	Amount string `bson:"amount" json:"amount"`
}

// ResponseResult is...
type ResponseResult struct {
	Error  string `json:"error"`
	Result string `json:"result"`
}
