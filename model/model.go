package model

import (
	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is...
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

// Vehicle is...
type Vehicle struct {
	VeicleID           primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	RegistrationNumber string             `bson:"registrationNumber" json:"registrationNumber"`
	VehicleClass       string             `bson:"vehicleClass" json:"vehicleClass"`
	UserID             string             `bson:"userId" json:"userId"`
}

// Payment is...
type Payment struct {
	PaymentID          primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	Days               interface{}        `bson:"days" json:"days"`
	VehicleReg         string             `bson:"vehicleReg" json:"vehicleReg"`
	Amount             int                `bson:"amount" json:"amount"`
	MpesaReceiptNumber string             `bson:"mpesaReceiptNumber" json:"mpesaReceiptNumber"`
	ResultCode         int                `bson:"resultCode" json:"resultCode"`
	ResultDesc         string             `bson:"resultDesc" json:"resultDesc"`
	TransactionDate    string             `bson:"transactionDate" json:"transactionDate"`
	PhoneNumber        string             `bson:"phoneNumber" json:"phoneNumber"`
	CheckoutRequestID  string             `bson:"checkoutRequestID" json:"checkoutRequestID"`
	IsSuccessful       bool               `bson:"isSuccessful" json:"isSuccessful"`
	UserID             string             `bson:"userId" json:"userId"`
}

// ResponseResult is...
type ResponseResult struct {
	Error  string `json:"error"`
	Result string `json:"result"`
}
