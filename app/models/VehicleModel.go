package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Vehicle struct {
	VeicleID           primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	RegistrationNumber string             `bson:"registrationNumber" json:"registrationNumber"`
	VehicleClass       string             `bson:"vehicleClass" json:"vehicleClass"`
	UserID             string             `bson:"userId" json:"userId"`
	IsWaitingClamp     bool               `bson:"isWaitingClamp" json:"isWaitingClamp"`
	IsClamped          bool               `bson:"isClamped" json:"isClamped"`
}
type Vehicles []Vehicle
