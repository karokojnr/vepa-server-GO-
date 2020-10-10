package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io/ioutil"
	"log"
	"time"
	model "vepa/app/models"
	"vepa/app/util"
)

func PaymentHandler(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("secret"), nil
	})
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		util.SendError(c, "Cannot get payment collection")
		return
	}
	var payment model.Payment
	err = c.Bind(&payment)
	if err != nil {
		util.SendError(c, "Error Getting Body")
		c.Abort()
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["id"].(string)
		payment.UserID = userID
		payment.IsSuccessful = false
		payment.PaymentID = primitive.NewObjectID()
		_, err = paymentCollection.InsertOne(ctx, payment)
		if err != nil {
			util.SendError(c, "Error Inserting Payment")
			c.Abort()
			return
		}
		util.Log("Payment added successfully..")
		c.JSON(200, gin.H{
			"message": "Payment added successfully..",
			"payment": &payment,
		})
		pID := payment.PaymentID.Hex()
		//INITIALIZE STK PUSH...
		GetPushHandler(userID, pID)
	}
	util.SendError(c, "You are not authorized!!!")
	c.Abort()
	return
}
func GetPushHandler(userID string, pID string) {
	util.Log("GetPushHandler Initialized...")
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
			util.Log("User not Found!")
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

func CallBackHandler(c *gin.Context) {
	log.Println("Callback called by M-pesa...")
	util.Log("Callback called by M-pesa...")
	log.Println("req body")
	ctx := context.TODO()
	userCollection, err := util.GetCollection("users")

	if err != nil {
		util.SendError(c, "Cannot get user collection")
		return
	}
	var bd interface{}
	rbody := c.Request.Body
	body, err := ioutil.ReadAll(rbody)
	err = json.Unmarshal(body, &bd)
	util.Log("Reading request body...")
	if err != nil {
		log.Println("Error")
		util.Log("Error parsing request:", err)
		return
	}
	userID := c.Request.URL.Query().Get("id")
	paymentID := c.Request.URL.Query().Get("paymentid")
	util.Log("Getting data from request...")
	util.Log("User ID:", userID, " Payment ID:", paymentID)
	idUser, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": idUser}
	var result model.User
	err = userCollection.FindOne(ctx, filter).Decode(&result)
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
	var paymentModel model.Payment
	resultCodeString := fmt.Sprintf("%v", resultCode)
	resultDesc := fmt.Sprintf("%v", rBody)

	if resultCodeString == string('0') {
		item = bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["CallbackMetadata"].(map[string]interface{})["Item"]
		mpesaReceiptNumber = item.([]interface{})[1].(map[string]interface{})["Value"]
		transactionDate = item.([]interface{})[3].(map[string]interface{})["Value"]
		//phoneNumber = item.([]interface{})[4].(map[string]interface{})["Value"]
		phoneNumber := "+" + result.PhoneNumber
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
			"phoneNumber":        phoneNumber,
			"checkoutRequestID":  checkoutRequestID,
			"isSuccessful":       true,
		}}
		err = paymentCollection.FindOneAndUpdate(ctx, paymentFilter, paymentUpdate).Decode(&paymentModel)
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
			"payment": paymentModel,
		})

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
		err = vehicleCollection.FindOneAndUpdate(ctx, vehicleFilter, vehicleUpdate).Decode(&vehicleModel)
		if err != nil {
			util.Log("Error fetching payment:", err.Error())
			fmt.Printf("error...")
			return

		}
		util.Log("vehicle paid")
		//------------------------------------------------------------//
		return
	}
	util.Log("Payment not successful")
	paymentModel.IsSuccessful = false
	//Send message(In case update was not successful)...
	util.SendNotifications(result.FCMToken, resultDesc)
	return

}
func UserPaymentsHandler(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("secret"), nil
	})
	var results []*model.Payment
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		util.SendError(c, "Cannot get payment collection")
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["id"].(string)
		filter := bson.M{"userId": userID, "isSuccessful": true}
		cur, err := paymentCollection.Find(ctx, filter)
		if err != nil {
			log.Fatal(err)
		}
		for cur.Next(ctx) {
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
		_ = cur.Close(ctx)
		c.JSON(200, gin.H{
			"payments": &results,
		})
		return
	}
	util.SendError(c, "You are not authorized!!!")
	c.Abort()
	return
}
func GetPaidDays(c *gin.Context) {
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		util.SendError(c, "Cannot get payment collection")
		return
	}
	vehicleReg := c.Param("vehicleReg")
	var results []*model.Payment
	filter := bson.M{"vehicleReg": vehicleReg, "isSuccessful": true}
	cur, err := paymentCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(ctx) {
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
	_ = cur.Close(ctx)
	c.JSON(200, gin.H{
		"payments": &results,
	})
	return
}
func VerificationHandler(c *gin.Context) {
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		util.SendError(c, "Cannot get payment collection")
		return
	}
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehicle collection")
		return
	}
	vehicleReg := c.Param("vehicleReg")
	var payment model.Payment
	var vehicle model.Vehicle
	err = vehicleCollection.FindOne(ctx, bson.M{"registrationNumber": vehicleReg}).Decode(&vehicle)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(200, gin.H{
				"result": "notfound",
			})
			return
		}

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
	err = paymentCollection.FindOne(ctx, filter).Decode(&payment)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(200, gin.H{
				"result": "unpaid",
			})
			return
		}
	}
	c.JSON(200, &payment)
	return
}
func UnpaidVehicleHistoryHandler(c *gin.Context) {
	ctx := context.TODO()
	paymentCollection, err := util.GetCollection("payments")
	if err != nil {
		util.SendError(c, "Cannot get payment collection")
		return
	}
	vehicleReg := c.Param("vehicleReg")

	var results []*model.Payment
	filter := bson.M{"vehicleReg": vehicleReg}
	cur, err := paymentCollection.Find(ctx, filter)
	if err != nil {
		util.SendError(c, "Error Getting All Clamped Vehicles")
		return

	}
	for cur.Next(ctx) {
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
	_ = cur.Close(ctx)
	c.JSON(200, gin.H{
		"result": &results,
	})
	return
}
