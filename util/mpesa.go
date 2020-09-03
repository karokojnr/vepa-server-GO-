package util

import (
	"vepa/model"
	// "vepa/util"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/AndroidStudyOpenSource/mpesa-api-go"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	appKey    = GoDotEnvVariable("MPESA_APP_KEY")
	appSecret = GoDotEnvVariable("MPESA_APP_SECRET")
)

//Push is...
func Push(userID string, pID string) {
	var rUser model.User
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	collection, err := GetUserCollection()
	if err != nil {
		log.Fatal(err)
	}
	err = collection.FindOne(context.TODO(), filter).Decode(&rUser)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			log.Println("User not Found!")
			// res.Result = "User not Found, Please try again later!"
			// json.NewEncoder(w).Encode(res)
			return
		}
	}

	svc, err := mpesa.New(appKey, appSecret, mpesa.SANDBOX)
	if err != nil {
		panic(err)
	}

	mres, err := svc.Simulation(mpesa.Express{
		BusinessShortCode: "174379",
		Password:          "MTc0Mzc5YmZiMjc5ZjlhYTliZGJjZjE1OGU5N2RkNzFhNDY3Y2QyZTBjODkzMDU5YjEwZjc4ZTZiNzJhZGExZWQyYzkxOTIwMjAwNDIxMTc1NTU1",
		Timestamp:         "20200421175555",
		TransactionType:   "CustomerPayBillOnline",
		Amount:            1,
		PartyA:            rUser.PhoneNumber,
		PartyB:            "174379",
		PhoneNumber:       rUser.PhoneNumber,
		CallBackURL:       "https://vepa-5c657.ew.r.appspot.com/rcb?id=" + userID + "&paymentid=" + pID,
		AccountReference:  "Vepa",
		TransactionDesc:   "Vepa Payment",
	})

	if err != nil {
		log.Println(err)
	}
	// log.Println(res)

	var mresMap map[string]interface{}
	errm := json.Unmarshal([]byte(mres), &mresMap)
	if errm != nil {
		log.Println("Error decoding response body")
		//panic(err)
	}
	rCode := mresMap["ResponseCode"]
	rCodeString := fmt.Sprintf("%v", rCode)
	rMessage := mresMap["ResponseDescription"]
	cMessage := mresMap["CustomerMessage"]
	log.Println(cMessage)
	//...

	// Send error message if error
	if rCodeString == string('0') {
		//// Proceed to STK Push
		//log.Println("rCode is zero...")
		//return
		//Do nothing...
		// runtime.Gosched()
		return

	}
	rMessageConv := fmt.Sprintf("%v", rMessage)
	//Send message...
	SendNotifications(rUser.FCMToken, rMessageConv)
	// wg.Done()
	return

}
