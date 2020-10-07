package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyokomi/emoji"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	// "sync"

	"vepa/model"
	"vepa/util"

	"github.com/AndroidStudyOpenSource/africastalking-go/sms"
	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func PaymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
	paymentCollection, err := util.GetCollection("payments")
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
		//TODO: Find vehicle
		_, err = paymentCollection.InsertOne(context.TODO(), payment)
		if err != nil {
			log.Println(err)
		}
		log.Println("Payment Added Successfully")
		res.Result = "Payment Added Successfully"
		json.NewEncoder(w).Encode(res)
		pID := payment.PaymentID.Hex()
		//INITIALIZE STK PUSH...
		//TODO: NOT being called sometimes
		GetPushHandler(userID, pID)
		// STK PUSH INITIALIZED
		//log.Println("stk push was initialized")
	}
	res.Error = "You are not Authorized!"
	json.NewEncoder(w).Encode(res)
	return

}

func GetPushHandler(userID string, pID string) {
	util.Log("GetPushHandler Initialized...")
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	// Get user to know the USER PHONE NUMBER
	var rUser model.User
	collection, err := util.GetCollection("users")
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
		CallBackURL:       "http://34.121.65.106:3500/rcb?id=" + userID + "&paymentid=" + pID, //CallBackHandler
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
	//log.Println(cMessage)
	util.Log(cMessage)

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

func CallBackHandler(w http.ResponseWriter, r *http.Request) {
	util.Log("Callback called by mpesa...")
	// Update Payment if payment was successful
	w.Header().Set("Content-Type", "application/json")
	var res model.ResponseResult
	var bd interface{}
	rbody := r.Body
	body, err := ioutil.ReadAll(rbody)
	err = json.Unmarshal(body, &bd)
	util.Log("Reading request body...")
	if err != nil {
		log.Println("Error")
		util.Log("Error parsing request:", err.Error())
		res.Result = "Unable to read request"
		json.NewEncoder(w).Encode(res)
		return
	}

	collection, err := util.GetCollection("users")
	if err != nil {
		log.Fatal(err)
	}
	//extract userId & paymentId
	_ = r.ParseForm() // Parses the request body
	userID := r.Form.Get("id")
	paymentID := r.Form.Get("paymentid")
	util.Log("Getting data from request...")
	util.Log("User ID:", userID, " Payment ID:", paymentID)
	idUser, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": idUser}
	var result model.User
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		util.Log("Error fetching user:", err.Error())
		if err.Error() == "mongo: no documents in result" {
			res.Result = "Something went wrong, Please try again later!"
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = "Error fetching user doc"
		json.NewEncoder(w).Encode(res)
		return
	}
	util.Log("User found:", result.Firstname, " Phone No:", result.PhoneNumber)

	util.Log("Reading result body...")
	resultCode := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["ResultCode"]
	rBody := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["ResultDesc"]
	checkoutRequestID := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["CheckoutRequestID"]

	util.Log("Result code:", resultCode, " Result Body:", rBody, " checkoutRequestID:", checkoutRequestID)

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
		util.Log("item:", item)
		util.Log("mpesaReceiptNumber:", mpesaReceiptNumber)
		util.Log("transactionDate:", transactionDate)
		util.Log("Fetching payment from db...")
		paymentCollection, err := util.GetCollection("payments")
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
			util.Log("Error fetching payment:", err.Error())
			fmt.Printf("error...")
			return

		}
		util.Log("Payment updated successfully...")
		//Send message(Payment was successful)...
		util.SendNotifications(result.FCMToken, resultDesc)
		res.Result = "Payment updated"
		json.NewEncoder(w).Encode(res)

		//-----Update is Waiting Clamp & isClamped in Vehicle-----
		vehicleCollection, err := util.GetCollection("vehicles")
		if err != nil {
			log.Fatal(err)
		}
		var vehicleModel model.Vehicle
		vehicleFilter := bson.M{"registrationNumber": paymentModel.VehicleReg}
		vehicleUpdate := bson.M{"$set": bson.M{
			"isWaitingClamp": false,
			"isClamped":      false,
		}}
		err = vehicleCollection.FindOneAndUpdate(context.TODO(), vehicleFilter, vehicleUpdate).Decode(&vehicleModel)
		if err != nil {
			util.Log("Error fetching payment:", err.Error())
			fmt.Printf("error...")
			return

		}
		util.Log("vehicle paid")
		//------------------------------------------------------------//
		return
	}
	util.Log("Payment no successful")
	paymentModel.IsSuccessful = false
	//Send message(In case update was not successful)...
	util.SendNotifications(result.FCMToken, resultDesc)
	return

}

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
	collection, err := util.GetCollection("payments")
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
	var results []*model.Payment
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		log.Println(err)
	}
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

}

func CheckVehicleClamp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(r)
	vehicleReg := params["vehicleReg"]
	//Check if vehicle has been clamped
	var vehicleModel model.Vehicle
	var res model.ResponseResult
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		log.Println(err)
	}
	vehicleFilter := bson.M{"registrationNumber": vehicleReg, "isClamped": true}
	err = vehicleCollection.FindOne(context.TODO(), vehicleFilter).Decode(&vehicleModel)
	if err != nil {
		json.NewEncoder(w).Encode(vehicleModel)
		return
	}
	res.Error = "clamped"
	json.NewEncoder(w).Encode(res)
	return

}

func VerificationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var params = mux.Vars(r)
	vehicleReg := params["vehicleReg"]

	var payment model.Payment
	// we get params with mux.
	var res model.ResponseResult

	var vehicle model.Vehicle
	vehicleCollection, err := util.GetCollection("vehicles")
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
	collection, err := util.GetCollection("payments")
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

func UnpaidVehicleHistoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(r)
	vehicleReg := params["vehicleReg"]
	var res model.ResponseResult
	var results []*model.Payment
	paymentCollection, err := util.GetCollection("payments")
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

func ClampVehicle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(r)
	vehicleReg := params["vehicleReg"]

	//TODO: Find vehicle
	var vehicle model.Vehicle
	//var res model.ResponseResult

	vehicleCollection, err := util.GetCollection("vehicles")
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
	vID := vehicle.VeicleID
	//vehicleID, _ := primitive.ObjectIDFromHex(vID)

	var user model.User
	userCollection, err := util.GetCollection("users")
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
	var vehicleModel model.Vehicle
	//var result1 model.Vehicle
	var res model.ResponseResult
	err = vehicleCollection.FindOne(context.TODO(), bson.M{"registrationNumber": vehicleReg}).Decode(&vehicleModel)
	if vehicleModel.IsWaitingClamp == true || vehicleModel.IsClamped == true {
		util.Log("vehicle is already  clamped")
		res.Error = "vehicle is already clamped"
		json.NewEncoder(w).Encode(res)
		return

	}
	//Send SMS - REPLACE Recipient and Message with REAL Values
	smsResponse, err := smsService.Send("", plus+userPhoneNumber, "Hello, Your have not paid for your vehicle("+vehicleReg+"). It will be clamped in 30 minutes incase you don't pay. Kindly make a payment now. ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(smsResponse)
	//------------isWaitingClamp==true----------//
	if vehicleModel.IsWaitingClamp == false || vehicleModel.IsClamped == false {
		vehicleFilter := bson.M{"_id": vID}
		vehicleUpdate := bson.M{"$set": bson.M{
			"isWaitingClamp": true,
		}}
		err = vehicleCollection.FindOneAndUpdate(context.TODO(), vehicleFilter, vehicleUpdate).Decode(&vehicleModel)
		if err != nil {
			util.Log("Error updating payment:", err.Error())
			fmt.Printf("error...")
			return

		}
		util.Log("isWaitingClamp == true")
		res.Result = "isWaitingClamp updated Successfully"
		json.NewEncoder(w).Encode(res)
		//return
	}
	//TODO:
	//--------------------------------------------//

	timerMessage := emoji.Sprint(":alarm_clock:")
	util.Log("Clamp timer started" + timerMessage)
	//clampTimer := time.NewTimer(30 * time.Second)
	//<-clampTimer.C
	time.Sleep(60 * time.Second)
	util.Log("Clamp timer ended" + timerMessage)
	//TODO: If paid dont proceed

	//------------isClamped==true----------//
	//var result2 model.Vehicle
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		log.Println(err)
	}
	var paymentModel model.Payment

	err = vehicleCollection.FindOne(context.TODO(), bson.M{"registrationNumber": vehicleReg}).Decode(&vehicleModel)
	currentTime := time.Now().Local()
	formatCurrentTime := currentTime.Format("2006-01-02")
	paymentFilter := bson.M{"vehicleReg": vehicleModel.RegistrationNumber, "days": formatCurrentTime, "isSuccessful": true}
	err = paymentCollection.FindOne(context.TODO(), paymentFilter).Decode(&paymentModel)
	if err != nil {
		log.Println(err)
		if vehicleModel.IsClamped == false {
			vehicleClampFilter := bson.M{"_id": vID}
			vehicleClampUpdate := bson.M{"$set": bson.M{
				"isClamped":      true,
				"isWaitingClamp": false,
			}}
			err = vehicleCollection.FindOneAndUpdate(context.TODO(), vehicleClampFilter, vehicleClampUpdate).Decode(&vehicleModel)
			if err != nil {
				util.Log("Error fetching payment:", err.Error())
				fmt.Printf("error...")
				return
			}
			util.Log("isClamped == true")
			res.Result = "isClamped updated Successfully"
			json.NewEncoder(w).Encode(res)
			//return
		}
	}
	vFilter := bson.M{"_id": vID}
	vUpdate := bson.M{"$set": bson.M{
		"isClamped":      false,
		"isWaitingClamp": false,
	}}
	err = vehicleCollection.FindOneAndUpdate(context.TODO(), vFilter, vUpdate).Decode(&vehicleModel)
	if err != nil {
		util.Log("Error fetching payment:", err.Error())
		fmt.Printf("error...")
		return
	}
	util.Log("Vehicle Parking Fee Paid, don't proceed to clamp")
	res.Result = "Paid, don't clamp"
	json.NewEncoder(w).Encode(res)

	//util.Log("vehicle has already been clamped")
	//res.Error = "vehicle has already been clamped"
	//json.NewEncoder(w).Encode(res)
	//--------------------------------------------//
	//TODO: Send notification to attendants
	//Send message to attendants(In case update was not successful)...
	util.SendNotifications("fi3ytpKGhRo:APA91bFqPPPFnpeQo2BRxB0NKTMfGxmaZNwX0XNu4NnJsz7inArbgrkDihHJF_om46NW2Bd-1pwHHZmOiV03s2hSZ_XLm2EkbxxOmwH9KukPaaZeq_0dSXe5giGCeD3s924XZDkMDfLv", "The vehicle has not yet been paid , Please clamp!")
	return

}
func ClearClampFee(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var clampFee model.ClampFee
	var res model.ResponseResult
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &clampFee)
	if err != nil {
		log.Println(err)
	}
	clampFeeCollection, err := util.GetCollection("clamps")
	if err != nil {
		log.Println(err)
	}
	var params = mux.Vars(r)
	//Get id from parameters
	userid := params["id"]
	//userID, _ := primitive.ObjectIDFromHex(userid)
	clampFee.UserID = userid
	clampFee.IsSuccessful = false
	clampFee.ClampFeeID = primitive.NewObjectID()
	_, err = clampFeeCollection.InsertOne(context.TODO(), clampFee)
	if err != nil {
		log.Println(err)
	}
	util.Log("Payment (amount), added successfully...")
	res.Result = "Payment Added Successfully"
	json.NewEncoder(w).Encode(res)
	cID := clampFee.ClampFeeID.Hex()

	log.Println(cID)
	//TODO: STK Push
	ClampPushHandler(userid, cID)

}
func ClampPushHandler(userID string, pID string) {
	util.Log("ClampPushHandler Initialized...")
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	// Get user to know the USER PHONE NUMBER
	var rUser model.User
	collection, err := util.GetCollection("users")
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
		CallBackURL:       "http://34.121.65.106:3500/clamprcb?id=" + userID + "&paymentID=" + pID, //ClampCallBackHandler
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
	//log.Println(cMessage)
	util.Log(cMessage)

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
func ClampCallBackHandler(w http.ResponseWriter, r *http.Request) {
	util.Log("ClampCallback called by Mpesa...")
	// Update Payment if payment was successful
	w.Header().Set("Content-Type", "application/json")
	var res model.ResponseResult
	var bd interface{}
	rbody := r.Body
	body, err := ioutil.ReadAll(rbody)
	err = json.Unmarshal(body, &bd)
	util.Log("Reading request body...")
	if err != nil {
		log.Println("Error")
		util.Log("Error parsing request:", err.Error())
		res.Result = "Unable to read request"
		json.NewEncoder(w).Encode(res)
		return
	}

	collection, err := util.GetCollection("users")
	if err != nil {
		log.Fatal(err)
	}
	//extract userId & paymentId
	_ = r.ParseForm() // Parses the request body
	userID := r.Form.Get("id")
	paymentID := r.Form.Get("paymentID")
	util.Log("Getting data from request...")
	util.Log("User ID:", userID, " Payment ID:", paymentID)
	idUser, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": idUser}
	var result model.User
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		util.Log("Error fetching user:", err.Error())
		if err.Error() == "mongo: no documents in result" {
			res.Result = "Something went wrong, Please try again later!"
			json.NewEncoder(w).Encode(res)
			return
		}
		res.Result = "Error fetching user doc"
		json.NewEncoder(w).Encode(res)
		return
	}
	util.Log("User found:", result.Firstname, " Phone No:", result.PhoneNumber)

	util.Log("Reading result body...")
	resultCode := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["ResultCode"]
	rBody := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["ResultDesc"]
	checkoutRequestID := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["CheckoutRequestID"]

	util.Log("Result code:", resultCode, " Result Body:", rBody, " checkoutRequestID:", checkoutRequestID)

	var item interface{}
	var mpesaReceiptNumber interface{}
	var transactionDate interface{}
	//var phoneNumber interface{}
	var clampPaymentModel model.ClampFee
	resultCodeString := fmt.Sprintf("%v", resultCode)
	resultDesc := fmt.Sprintf("%v", rBody)

	var vehicleModel model.Vehicle
	//-----Update is Waiting Clamp & isClamped in Vehicle-----
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		log.Fatal(err)
	}
	if resultCodeString == string('0') {
		item = bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["CallbackMetadata"].(map[string]interface{})["Item"]
		mpesaReceiptNumber = item.([]interface{})[1].(map[string]interface{})["Value"]
		transactionDate = item.([]interface{})[3].(map[string]interface{})["Value"]
		//phoneNumber = item.([]interface{})[4].(map[string]interface{})["Value"]
		// phoneNumber = result.PhoneNumber
		util.Log("item:", item)
		util.Log("mpesaReceiptNumber:", mpesaReceiptNumber)
		util.Log("transactionDate:", transactionDate)
		util.Log("Fetching payment from db...")
		clampFeeCollection, err := util.GetCollection("clamps")
		if err != nil {
			log.Fatal(err)
		}
		pid, _ := primitive.ObjectIDFromHex(paymentID)
		clampPaymentFilter := bson.M{"_id": pid}
		clampPaymentUpdate := bson.M{"$set": bson.M{
			"amount":             1,
			"mpesaReceiptNumber": mpesaReceiptNumber,
			"resultCode":         resultCode,
			"resultDesc":         resultDesc,
			"transactionDate":    transactionDate,
			//"phoneNumber":        phoneNumber,
			"checkoutRequestID": checkoutRequestID,
			"isSuccessful":      true,
		}}
		err = clampFeeCollection.FindOneAndUpdate(context.TODO(), clampPaymentFilter, clampPaymentUpdate).Decode(&clampPaymentModel)
		if err != nil {
			util.Log("Error fetching payment:", err.Error())
			fmt.Printf("error...")
			return

		}
		util.Log("Payment updated successfully...")
		//Send message(Payment was successful)...
		util.SendNotifications(result.FCMToken, resultDesc)
		res.Result = "Payment updated"
		json.NewEncoder(w).Encode(res)

		//Set isClamped == false
		util.Log("Vehicle Reg - ", clampPaymentModel.VehicleReg)
		vehicleFilter := bson.M{"registrationNumber": clampPaymentModel.VehicleReg}
		vehicleUpdate := bson.M{"$set": bson.M{
			"isWaitingClamp": false,
			"isClamped":      false,
		}}
		err = vehicleCollection.FindOneAndUpdate(context.TODO(), vehicleFilter, vehicleUpdate).Decode(&vehicleModel)
		if err != nil {
			util.Log("Error updating vehicle:", err.Error())
			fmt.Printf("error...")
			return

		}
		util.Log("Clamp fee cleared...")
		return
	}
	util.Log("Payment not successful")
	clampPaymentModel.IsSuccessful = false
	//Set isClamped to true
	//-----Update is Waiting Clamp & isClamped in Vehicle-----
	vehicleFilter := bson.M{"registrationNumber": clampPaymentModel.VehicleReg}
	vehicleUpdate := bson.M{"$set": bson.M{
		"isWaitingClamp": false,
		"isClamped":      true,
	}}
	var vModel model.Vehicle
	err = vehicleCollection.FindOneAndUpdate(context.TODO(), vehicleFilter, vehicleUpdate).Decode(&vModel)
	if err != nil {
		util.Log("Error updating vehicle:", err.Error())
		fmt.Printf("error...")
		return

	}
	util.Log("vehicle clamp fee not paid")
	//Send message(In case update was not successful)...
	util.SendNotifications(result.FCMToken, resultDesc)
	return

}
