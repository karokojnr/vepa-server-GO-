package model

import "go.mongodb.org/mongo-driver/bson/primitive"
type User struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
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
