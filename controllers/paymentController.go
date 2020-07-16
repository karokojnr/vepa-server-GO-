package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"time"
	"vepa/model"
	"vepa/util"
	//"vepa/util/notificationsService"

	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
// PaymentHandler is...
func PaymentHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-TYpe", "application/json")
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("secret"), nil
	})
	var payment model.Payment
	var res model.ResponseResult

	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &payment)
	collection, err := util.GetPaymentCollection()
	if err != nil {
		res.Error = "Error, Try Again Later"
		json.NewEncoder(w).Encode(res)
		return
	}
	// var result model.Payment
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["id"].(string)
		payment.UserID = userID
		payment.IsSuccessful = false
		payment.PaymentID = primitive.NewObjectID()

		fmt.Println("Payment Handeler Used ID:")
		log.Println(userID)
		_, err = collection.InsertOne(context.TODO(), payment)
		if err != nil {
			res.Error = "Error While Making Payment, Try Again"
			json.NewEncoder(w).Encode(res)
			return
		}

		// pID:= payment.PaymentID

		res.Result = "Payment Added Successfully"
		json.NewEncoder(w).Encode(res)
		pID := payment.PaymentID.Hex()

		var (
			appKey    = util.GoDotEnvVariable("MPESA_APP_KEY")
			appSecret = util.GoDotEnvVariable("MPESA_APP_SECRET")
		)
		svc, err := mpesa.New(appKey, appSecret, mpesa.SANDBOX)
		if err != nil {
			panic(err)
		}

		mres, err := svc.Simulation(mpesa.Express{
			BusinessShortCode: "174379",
			Password:          util.GoDotEnvVariable("MPESA_PASSWORD"),
			Timestamp:         "20200421175555",
			TransactionType:   "CustomerPayBillOnline",
			Amount:            1,
			PartyA:            "254799338805",
			PartyB:            "174379",
			PhoneNumber:       "254799338805",
			CallBackURL:       "https://vepa-5c657.ew.r.appspot.com/rcb?id=" + userID + "&paymentid=" + pID,
			AccountReference:  "Vepa",
			TransactionDesc:   "Vepa Payment",
		})

		if err != nil {
			log.Println("err")
		}
		var mresMap map[string]interface{}
		errm := json.Unmarshal([]byte(mres), &mresMap)

		if errm != nil {
			panic(err)
		}
		rCode := mresMap["ResponseCode"]
		rMessage := mresMap["ResponseDescription"]
		cMessage := mresMap["CustomerMessage"]
		log.Println(cMessage)
		// Send error message if error
		if rCode != 0 {

			//Send message...
			id, _ := primitive.ObjectIDFromHex(userID)
			filter := bson.M{"_id": id}

			var rUser model.User
			err = collection.FindOne(context.TODO(), filter).Decode(&rUser)
			if err != nil {
				if err.Error() == "mongo: no documents in result" {
					res.Result = "Something went wrong, Please try again later!"
					json.NewEncoder(w).Encode(res)
					return
				}
			}
			rMessageConv := fmt.Sprintf("%v", rMessage)
			//Send message...
			util.SendNotifications(rUser.FCMToken, rMessageConv)
			return
		}
		log.Println(mres)
		return
	}
	res.Error = err.Error()
	json.NewEncoder(w).Encode(res)
	return

}

// CallBackHandler is...
func CallBackHandler(w http.ResponseWriter, r *http.Request) {
	//defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("IM INSIDE CALBACK")
	var bd interface{}
	rbody := r.Body
	body, err := ioutil.ReadAll(rbody)
	err = json.Unmarshal(body, &bd)
	if err != nil {
		log.Println("eRROR")
	}
	if err != nil {
		panic(err)
	}
	var res model.ResponseResult
	collection, err := util.GetUserCollection()
	if err != nil {
		log.Fatal(err)
	}
	//extract userId
	_ = r.ParseForm() // Parses the request body
	userID := r.Form.Get("id")
	paymentID := r.Form.Get("paymentid")
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}

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
	fmt.Println("RBODY TYPE:")
	fmt.Println(reflect.TypeOf(rbody))
	var item interface{}
	var mpesaReceiptNumber interface{}
	var transactionDate interface{}
	var phoneNumber interface{}

	checkoutRequestID := bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["CheckoutRequestID"]

	if resultCode == 0 {
		item = bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["CallbackMetadata"].(map[string]interface{})["Item"]
		mpesaReceiptNumber = item.([]interface{})[1].(map[string]interface{})["Value"]
		transactionDate = item.([]interface{})[3].(map[string]interface{})["Value"]
		phoneNumber = item.([]interface{})[4].(map[string]interface{})["Value"]
	}

	paymentCollection, err := util.GetPaymentCollection()
	if err != nil {
		log.Fatal(err)
	}
	pid, _ := primitive.ObjectIDFromHex(paymentID)
	paymentFilter := bson.M{"_id": pid}
	var paymenModel model.Payment
	fmt.Println("Payment ID")
	log.Println(paymentID)
	_ = json.NewDecoder(r.Body).Decode(&paymenModel)
	if resultCode != 0 {
		paymenModel.IsSuccessful = false
	}
	paymentUpdate := bson.M{"$set": bson.M{
		"mpesaReceiptNumber": mpesaReceiptNumber,
		"resultCode":         resultCode,
		"resultDesc":         rBody,
		"transactionDate":    transactionDate,
		"phoneNumber":        phoneNumber,
		"checkoutRequestID":  checkoutRequestID,
		"isSuccessful":       true,
	}}
	errp := paymentCollection.FindOneAndUpdate(context.TODO(), paymentFilter, paymentUpdate).Decode(&paymenModel)
	if errp != nil {
		fmt.Printf("error...")
		return

	}
	rBodyConv := fmt.Sprintf("%v", rBody)
	//Send message...
	time.Sleep(2 * time.Second)
	util.SendNotifications(result.FCMToken, rBodyConv)
	res.Result = "Payment updated"
	json.NewEncoder(w).Encode(res)
	return
}

// UserPaymentsHandler is...
func UserPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-TYpe", "application/json")
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("secret"), nil
	})
	var res model.ResponseResult
	var results []*model.Payment
	// errr := json.Unmarshal([]byte(s), &results)
	// if errr != nil {
	// 	log.Println("Unmarshall error")
	// }
	collection, err := util.GetPaymentCollection()
	if err != nil {
		res.Error = "Error, Try Again Later"
		json.NewEncoder(w).Encode(res)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["id"].(string)
		// userID, _ := primitive.ObjectIDFromHex(id)
		filter := bson.M{"userId": userID}
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
	res.Error = err.Error()
	json.NewEncoder(w).Encode(res)
	return
}
