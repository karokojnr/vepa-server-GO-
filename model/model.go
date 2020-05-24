package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID          primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Firstname   string             `json:"firstName"`
	Lastname    string             `json:"lastName"`
	Email       string             `json:"email"`
	IDNumber    string             `json:"idNumber"`
	PhoneNumber string             `json:"phoneNumber"`
	Password    string             `json:"password"`
	Token       string             `json:"token"`
}
type Vehicle struct {
	RegistrationNumber string `bson:"registrationNumber" json:"registrationNumber"`
	VehicleClass       string `bson:"vehicleClass" json:"vehicleClass"`
	VehicleModel       string `bson:"vehicleModel" json:"vehicleModel"`
	UserID             string `bson:"userId" json:"userId"`
}

type ResponseResult struct {
	Error  string `json:"error"`
	Result string `json:"result"`
}
