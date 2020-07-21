package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Payment is...
type Payment struct {
	PaymentID          primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	Days               []interface{}      `bson:"days" json:"days"`
	VehicleReg         string             `bson:"vehicleReg" json:"vehicleReg"`
	Amount             int                `bson:"amount" json:"amount"`
	MpesaReceiptNumber string             `bson:"mpesaReceiptNumber" json:"mpesaReceiptNumber"`
	ResultCode         interface{}                `bson:"resultCode" json:"resultCode"`
	ResultDesc         string             `bson:"resultDesc" json:"resultDesc"`
	TransactionDate    int                `bson:"transactionDate" json:"transactionDate"`
	PhoneNumber        int                `bson:"phoneNumber" json:"phoneNumber"`
	CheckoutRequestID  string             `bson:"checkoutRequestID" json:"checkoutRequestID"`
	IsSuccessful       bool               `bson:"isSuccessful" json:"isSuccessful"`
	UserID             string             `bson:"userId" json:"userId"`
}
