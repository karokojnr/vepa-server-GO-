package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AndroidStudyOpenSource/africastalking-go/sms"
	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	"github.com/gin-gonic/gin"
	"github.com/kyokomi/emoji"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io/ioutil"
	"log"
	"time"
	model "vepa/app/models"
	"vepa/app/util"
)

func ClampVehicleHandler(c *gin.Context) {
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		util.SendError(c, "Cannot get clamp fee collection")
		return
	}
	userCollection, err := util.GetCollection("users")
	if err != nil {
		util.SendError(c, "Cannot get user collection")
		return
	}
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehicle collection")
		return
	}
	vehicleReg := c.Param("vehicleReg")
	var vehicle model.Vehicle
	filter := bson.M{"registrationNumber": vehicleReg}
	err = vehicleCollection.FindOne(ctx, filter).Decode(&vehicle)
	if err != nil {
		log.Println(err)
	}
	//Find userID to get the phone number
	uID := vehicle.UserID
	userID, _ := primitive.ObjectIDFromHex(uID)
	vID := vehicle.VeicleID

	var user model.User
	err = userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		log.Println(err)
	}
	userPhoneNumber := user.PhoneNumber
	//Test if phone number is available
	log.Println("---Phone Number---")
	log.Println(userPhoneNumber)
	util.Log("User phone number", userPhoneNumber)

	var (
		username = "karokojnr"                                        //Your Africa's Talking Username
		apiKey   = util.GoDotEnvVariable("AFRICA_IS_TALKING_API_KEY") //Production or Sandbox API Key
		env      = "production"                                       // Choose either Sandbox or Production
	)
	//Call the Gateway, and pass the constants here!
	smsService := sms.NewService(username, apiKey, env)
	plus := "+"
	var vehicleModel model.Vehicle
	err = vehicleCollection.FindOne(ctx, bson.M{"registrationNumber": vehicleReg}).Decode(&vehicleModel)
	if vehicleModel.IsWaitingClamp == true || vehicleModel.IsClamped == true {
		util.Log("vehicle is already  clamped")
		util.SendError(c, "vehicle is already clamped")
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
			util.SendError(c, "Error Updating Vehicle")
			return
		}
		util.Log("isWaitingClamp == true")
		c.JSON(200, gin.H{
			"message": "isWaitingClamp updated Successfully --> true",
		})
	}
	//Set Timer
	timerMessage := emoji.Sprint(":alarm_clock:")
	util.Log("Clamp timer started" + timerMessage)
	time.Sleep(60 * time.Second)
	util.Log("Clamp timer ended" + timerMessage)
	//Timer complete

	var paymentModel model.Payment
	err = vehicleCollection.FindOne(ctx, bson.M{"registrationNumber": vehicleReg}).Decode(&vehicleModel)
	currentTime := time.Now().Local()
	formatCurrentTime := currentTime.Format("2006-01-02")
	paymentFilter := bson.M{"vehicleReg": vehicleModel.RegistrationNumber, "days": formatCurrentTime, "isSuccessful": true}
	err = paymentCollection.FindOne(ctx, paymentFilter).Decode(&paymentModel)
	if err != nil {
		log.Println(err)
		if err.Error() == "mongo: no documents in result" {
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
					util.SendError(c, "Error Updating Vehicle")
					return
				}
				util.Log("isClamped == true")
				c.JSON(200, gin.H{
					"message": "isClamped updated Successfully --> true",
				})
				return
			}

		}
	}
	vFilter := bson.M{"_id": vID}
	vUpdate := bson.M{"$set": bson.M{
		"isClamped":      false,
		"isWaitingClamp": false,
	}}
	err = vehicleCollection.FindOneAndUpdate(context.TODO(), vFilter, vUpdate).Decode(&vehicleModel)
	if err != nil {
		util.Log("Error updating payment:", err.Error())
		fmt.Printf("error...")
		return
	}
	util.Log("Vehicle Parking Fee Paid, don't proceed to clamp")
	c.JSON(200, gin.H{
		"message": "Paid, don't clamp",
	})
	util.SendNotifications("fi3ytpKGhRo:APA91bFqPPPFnpeQo2BRxB0NKTMfGxmaZNwX0XNu4NnJsz7inArbgrkDihHJF_om46NW2Bd-1pwHHZmOiV03s2hSZ_XLm2EkbxxOmwH9KukPaaZeq_0dSXe5giGCeD3s924XZDkMDfLv", "The vehicle has not yet been paid , Please clamp!")
	return
}

func ClearClampFeeHandler(c *gin.Context) {
	ctx := context.TODO()
	clampFeeCollection, err := util.GetCollection("clamps")
	if err != nil {
		util.SendError(c, "Cannot get clamp fee collection")
		return
	}
	var clampFee model.ClampFee
	err = c.Bind(&clampFee)
	if err != nil {
		util.SendError(c, "Error Getting Body")
		c.Abort()
		return
	}
	userID := c.Param("id")
	clampFee.UserID = userID
	clampFee.IsSuccessful = false
	clampFee.ClampFeeID = primitive.NewObjectID()
	_, err = clampFeeCollection.InsertOne(ctx, clampFee)
	if err != nil {
		util.SendError(c, "Error Inserting Clamp Fee Payment")
		c.Abort()
		return
	}
	util.Log("Payment (amount), added successfully...")
	c.JSON(200, gin.H{
		"message":   "Success Insert Vehicle",
		"clamp fee": &clampFee,
	})
	cID := clampFee.ClampFeeID.Hex()
	log.Println(cID)
	//STK Push
	ClampPushHandler(userID, cID)

}
func ClampPushHandler(userID string, cID string) {
	util.Log("ClampPushHandler Initialized...")
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	// Get user to know the USER PHONE NUMBER
	userCollection, err := util.GetCollection("users")
	if err != nil {
		log.Fatal(err)
		return
	}
	var rUser model.User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&rUser)
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
		CallBackURL:       "http://34.121.65.106:3500/clamprcb?id=" + userID + "&paymentID=" + cID, //ClampCallBackHandler
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

func ClampCallBackHandler(c *gin.Context) {
	util.Log("ClampCallback called by Mpesa...")
	// Update Payment if payment was successful
	var bd interface{}
	rbody := c.Request.Body
	body, err := ioutil.ReadAll(rbody)
	err = json.Unmarshal(body, &bd)
	util.Log("Reading request body...")
	if err != nil {
		log.Println("Error")
		util.Log("Error parsing request:", err.Error())
		return
	}

	collection, err := util.GetCollection("users")
	if err != nil {
		log.Fatal(err)
	}
	//extract userId & paymentId
	userID := c.Request.URL.Query().Get("id")
	paymentID := c.Request.URL.Query().Get("paymentid")
	util.Log("Getting data from request...")
	util.Log("User ID:", userID, " Payment ID:", paymentID)
	idUser, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": idUser}
	var result model.User
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		util.Log("Error fetching user:", err.Error())
		if err.Error() == "mongo: no documents in result" {
			util.SendError(c, "User account was not found")
			c.Abort()
			return
		}
		util.SendError(c, "Error fetching user doc")
		c.Abort()
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
		phoneNumber := result.PhoneNumber
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
			"phoneNumber":        phoneNumber,
			"checkoutRequestID":  checkoutRequestID,
			"isSuccessful":       true,
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
		c.JSON(200, gin.H{
			"message": "Payment Updated",
			"payment": clampPaymentModel,
		})

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
	util.Log("Payment not successful...")
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
