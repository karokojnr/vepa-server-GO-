package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	// "sync"

	"vepa/model"
	"vepa/util"

	"github.com/AndroidStudyOpenSource/africastalking-go/sms"
	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	"github.com/gorilla/mux"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentHandler is...
func PaymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-TYpe", "application/json")
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("secret"), nil
	})
	var payment model.Payment
	var res model.ResponseResult
	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &payment)
	paymentCollection, err := util.GetPaymentCollection()
	if err != nil {
		res.Error = "Error, Try Again Later"
		json.NewEncoder(w).Encode(res)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["id"].(string)
		payment.UserID = userID
		payment.IsSuccessful = false
		payment.PaymentID = primitive.NewObjectID()
		_, err = paymentCollection.InsertOne(context.TODO(), payment)
		if err != nil {
			log.Println(err)
		}
		log.Println("Payment Added Successfully")
		res.Result = "Payment Added Successfully"
		json.NewEncoder(w).Encode(res)

		pID := payment.PaymentID.Hex()
		//INITIALIZE STK PUSH...
		GetPushHandler(userID, pID)
		// STK PUSH INITIALIZED
		log.Println("stk push was initialized")
	}
	res.Error = "You are not Authorized!"
	json.NewEncoder(w).Encode(res)
	return

}

//GetPushHandler is..
func GetPushHandler(userID string, pID string) {
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	// Get user to know the USER PHONE NUMBER
	var rUser model.User
	collection, err := util.GetUserCollection()
	if err != nil {
		log.Fatal(err)
	}
	err = collection.FindOne(context.TODO(), filter).Decode(&rUser)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			log.Println("User not Found!")
			return
		}
	}
	//Initialize STK Push
	var (
		appKey    = util.GoDotEnvVariable("MPESA_APP_KEY")
		appSecret = util.GoDotEnvVariable("MPESA_APP_SECRET")
	)
	svc, err := mpesa.New(appKey, appSecret, mpesa.SANDBOX)
	if err != nil {
		log.Println(err)
	}
	mres, err := svc.Simulation(mpesa.Express{
		BusinessShortCode: "174379",
		Password:          util.GoDotEnvVariable("MPESA_PASSWORD"),
		Timestamp:         "20200421175555",
		TransactionType:   "CustomerPayBillOnline",
		Amount:            1,
		PartyA:            rUser.PhoneNumber,
		PartyB:            "174379",
		PhoneNumber:       rUser.PhoneNumber,
		CallBackURL:       "https://vepa-5c657.ew.r.appspot.com/rcb?id=" + userID + "&paymentid=" + pID, //CallBackHandler
		AccountReference:  "Vepa",
		TransactionDesc:   "Vepa Payment",
	})
	if err != nil {
		log.Println("STK Push not sent")
	}

	var mresMap map[string]interface{}
	errm := json.Unmarshal([]byte(mres), &mresMap)
	if errm != nil {
		log.Println("Error decoding response body")
	}
	rCode := mresMap["ResponseCode"]
	rCodeString := fmt.Sprintf("%v", rCode)
	rMessage := mresMap["ResponseDescription"]
	cMessage := mresMap["CustomerMessage"]
	log.Println(cMessage)

	// Send error message(notification) if rCode != 0
	if rCodeString == string('0') {
		//// Proceed to STK Push
		return

	}
	rMessageConv := fmt.Sprintf("%v", rMessage)
	//Send message...
	util.SendNotifications(rUser.FCMToken, rMessageConv)
	return

}

// CallBackHandler is...
func CallBackHandler(w http.ResponseWriter, r *http.Request) {
	// Update Payment if payment was successful
	w.Header().Set("Content-Type", "application/json")
	var bd interface{}
	rbody := r.Body
	body, err := ioutil.ReadAll(rbody)
	err = json.Unmarshal(body, &bd)
	if err != nil {
		log.Println("eRROR")
	}
	var res model.ResponseResult
	collection, err := util.GetUserCollection()
	if err != nil {
		log.Fatal(err)
	}
	//extract userId & paymentId
	_ = r.ParseForm() // Parses the request body
	userID := r.Form.Get("id")
	paymentID := r.Form.Get("paymentid")
	idUser, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": idUser}
	var result model.User
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			res.Result = "Something went wrong, Please try again later!"
			json.NewEncoder(w).Encode(res)
			return
		}
	}
	resultCode := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["ResultCode"]
	rBody := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["ResultDesc"]
	checkoutRequestID := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["CheckoutRequestID"]

	var item interface{}
	var mpesaReceiptNumber interface{}
	var transactionDate interface{}
	//var phoneNumber interface{}
	var paymentModel model.Payment
	resultCodeString := fmt.Sprintf("%v", resultCode)
	resultDesc := fmt.Sprintf("%v", rBody)

	if resultCodeString == string('0') {
		item = bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["CallbackMetadata"].(map[string]interface{})["Item"]
		mpesaReceiptNumber = item.([]interface{})[1].(map[string]interface{})["Value"]
		transactionDate = item.([]interface{})[3].(map[string]interface{})["Value"]
		//phoneNumber = item.([]interface{})[4].(map[string]interface{})["Value"]
		// phoneNumber = result.PhoneNumber

		paymentCollection, err := util.GetPaymentCollection()
		if err != nil {
			log.Fatal(err)
		}
		pid, _ := primitive.ObjectIDFromHex(paymentID)
		paymentFilter := bson.M{"_id": pid}
		paymentUpdate := bson.M{"$set": bson.M{
			"amount":             1,
			"mpesaReceiptNumber": mpesaReceiptNumber,
			"resultCode":         resultCode,
			"resultDesc":         resultDesc,
			"transactionDate":    transactionDate,
			//"phoneNumber":        phoneNumber,
			"checkoutRequestID": checkoutRequestID,
			"isSuccessful":      true,
		}}
		err = paymentCollection.FindOneAndUpdate(context.TODO(), paymentFilter, paymentUpdate).Decode(&paymentModel)
		if err != nil {
			fmt.Printf("error...")
			return

		}
		//Send message(Payment was successful)...
		util.SendNotifications(result.FCMToken, resultDesc)
		res.Result = "Payment updated"
		json.NewEncoder(w).Encode(res)
		return
	}
	paymentModel.IsSuccessful = false
	//Send message(In case update was not successful)...
	util.SendNotifications(result.FCMToken, resultDesc)
	return

}

// UserPaymentsHandler is...
func UserPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("secret"), nil
	})
	var res model.ResponseResult
	var results []*model.Payment
	collection, err := util.GetPaymentCollection()
	if err != nil {
		res.Error = "Error, Try Again Later"
		json.NewEncoder(w).Encode(res)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["id"].(string)
		//userID, _ := primitive.ObjectIDFromHex(id)
		filter := bson.M{"userId": userID, "isSuccessful": true}
		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			log.Fatal(err)
		}
		for cur.Next(context.TODO()) {
			var elem model.Payment
			err := cur.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}
			results = append(results, &elem)

		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}
		_ = cur.Close(context.TODO())
		json.NewEncoder(w).Encode(results)
		return
	}
	res.Error = "You are not Authorized!"
	json.NewEncoder(w).Encode(res)
	return
}

// GetPaidDays given vehicle registration number...
func GetPaidDays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//tokenString := r.Header.Get("Authorization")
	//token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	//	// Don't forget to validate the alg is what you expect:
	//	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
	//		return nil, fmt.Errorf("unexpected signing method")
	//	}
	//	return []byte("secret"), nil
	//})
	var params = mux.Vars(r)
	vehicleReg := params["vehicleReg"]
	//var res model.ResponseResult
	var results []*model.Payment
	paymentCollection, err := util.GetPaymentCollection()
	if err != nil {
		log.Println(err)
	}

	//if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
	//userID := claims["id"].(string)
	filter := bson.M{"vehicleReg": vehicleReg, "isSuccessful": true}
	cur, err := paymentCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		var elem model.Payment
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	_ = cur.Close(context.TODO())
	json.NewEncoder(w).Encode(results)
	return
	//}
	//res.Error = err.Error()
	//json.NewEncoder(w).Encode(res)
	//return

}

//VerificationHandler is...
func VerificationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var params = mux.Vars(r)
	vehicleReg := params["vehicleReg"]

	var payment model.Payment
	// we get params with mux.
	var res model.ResponseResult

	var vehicle model.Vehicle
	vehicleCollection, err := util.GetVehicleCollection()
	if err != nil {
		log.Fatal(err)
	}
	err = vehicleCollection.FindOne(context.TODO(), bson.M{"registrationNumber": vehicleReg}).Decode(&vehicle)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			res.Result = "notfound"
			json.NewEncoder(w).Encode(res)
			return
		}
	}
	collection, err := util.GetPaymentCollection()
	if err != nil {
		log.Fatal(err)
	}
	currentTime := time.Now().Local()
	formatCurrentTime := currentTime.Format("2006-01-02")

	for i := range payment.Days {
		if payment.Days[i] == formatCurrentTime {
			fmt.Println("Found")
			// Found!
		}
	}
	log.Println("---Payment Days---")
	log.Println(payment.Days)
	// We create filter. If it is unnecessary to sort data for you, you can use bson.M{}
	filter := bson.M{"vehicleReg": vehicleReg, "days": formatCurrentTime, "isSuccessful": true}
	err = collection.FindOne(context.TODO(), filter).Decode(&payment)

	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			res.Result = "unpaid"
			json.NewEncoder(w).Encode(res)
			return
		}
	}
	json.NewEncoder(w).Encode(payment)

}

//UnpaidVehicleHistoryHandler is...
func UnpaidVehicleHistoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(r)
	vehicleReg := params["vehicleReg"]
	var res model.ResponseResult
	var results []*model.Payment
	paymentCollection, err := util.GetPaymentCollection()
	if err != nil {
		res.Error = "Error, Try again later"
		json.NewEncoder(w).Encode(res)
		return
	}
	filter := bson.M{"vehicleReg": vehicleReg}
	cur, err := paymentCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)

	}
	for cur.Next(context.TODO()) {
		var elem model.Payment
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	_ = cur.Close(context.TODO())
	json.NewEncoder(w).Encode(results)
	return

}

//ClampVehicle is...
func ClampVehicle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(r)
	vehicleReg := params["vehicleReg"]

	//TODO: Find vehicle
	var vehicle model.Vehicle
	//var res model.ResponseResult

	vehicleCollection, err := util.GetVehicleCollection()
	if err != nil {
		log.Println(err)
	}
	filter := bson.M{"registrationNumber": vehicleReg}
	err = vehicleCollection.FindOne(context.TODO(), filter).Decode(&vehicle)
	if err != nil {
		log.Println(err)
	}
	//Find userID to get the phone number
	uID := vehicle.UserID
	userID, _ := primitive.ObjectIDFromHex(uID)

	var user model.User
	userCollection, err := util.GetUserCollection()
	if err != nil {
		log.Println(err)
	}
	err = userCollection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		log.Println(err)
	}
	userPhoneNumber := user.PhoneNumber
	//Test if phone number is available
	log.Println("---Phone Number---")
	log.Println(userPhoneNumber)
	var (
		username = "karokojnr"                                        //Your Africa's Talking Username
		apiKey   = util.GoDotEnvVariable("AFRICA_IS_TALKING_API_KEY") //Production or Sandbox API Key
		env      = "production"                                       // Choose either Sandbox or Production
	)
	//Call the Gateway, and pass the constants here!
	smsService := sms.NewService(username, apiKey, env)
	plus := "+"

	//Send SMS - REPLACE Recipient and Message with REAL Values
	smsResponse, err := smsService.Send("", plus+userPhoneNumber, "Hello, Your have not paid for your vehicle. It will be clamped in 30 minutes incase you don't pay. Kindly make a payment now. ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(smsResponse)

}
