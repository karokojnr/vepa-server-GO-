package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type ClampDetails struct {
	ClampDetailID   primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	VehicleReg      string             `bson:"vehicleReg" json:"vehicleReg"`
	AttendantID     string             `bson:"attendantId" json:"attendantId"`
	IsCarRegistered bool               `bson:"isCarRegistered" json:"isCarRegistered"`
	ClampDate       interface{}        `bson:"clampDate" json:"clampDate"`
}
