package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"vepa/model"
	"vepa/util"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentHandler is...
func PaymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-TYpe", "application/json")
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
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
		//userID, _ := primitive.Hex(payment.PaymentID)
		var result model.Payment
		err = paymentCollection.FindOne(context.TODO(), bson.M{"_id": payment.PaymentID, "days": payment.Days, "isSuccessful": true}).Decode(&result)

		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				//Insert data into Payment collection
				_, err = paymentCollection.InsertOne(context.TODO(), payment)
				if err != nil {
					res.Error = "Error While Making Payment, Try Again"
					json.NewEncoder(w).Encode(res)
					return
				}
				log.Println("Payment Added Successfully")
				res.Result = "Payment Added Successfully"
				json.NewEncoder(w).Encode(res)
				pID := payment.PaymentID.Hex()
				//
				//STK PUSH...
				//GetPushHandler(w, userID, pID)
				//
				resp, err := http.Get("https://vepa-5c657.ew.r.appspot.com/paymentPush?userID=" + userID + "&pID=" + pID)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				//body, err := ioutil.ReadAll(resp.Body)
				//if err != nil {
				//	panic(err)
				//}
				//fmt.Printf("%s", body)

			}
			res.Error = "Kindly Choose a day that hasn't been paid for!"
			json.NewEncoder(w).Encode(res)
			return
		}

	}
	res.Error = "You are not Authorized!"
	json.NewEncoder(w).Encode(res)
	return

}
//GetPushHandler is..
func GetPushHandler(w http.ResponseWriter, r *http.Request, ) {
	//extract userId & paymentId
	_ = r.ParseForm() // Parses the request body
	userID := r.Form.Get("userID")
	pID := r.Form.Get("pID")
	// userID string, pID string
	var res model.ResponseResult
	//Get current logged in user details
	var rUser model.User
	id, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": id}
	collection, err := util.GetUserCollection()
	if err != nil {
		log.Fatal(err)
	}
	err = collection.FindOne(context.TODO(), filter).Decode(&rUser)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			log.Println("User not Found!")
			res.Result = "User not Found, Please try again later!"
			json.NewEncoder(w).Encode(res)
			return
		}
	}
	//STK Push
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
		PartyA:            rUser.PhoneNumber,
		PartyB:            "174379",
		PhoneNumber:       rUser.PhoneNumber,
		CallBackURL:       "https://vepa-5c657.ew.r.appspot.com/rcb?id=" + userID + "&paymentid=" + pID,
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

	} else {
		rMessageConv := fmt.Sprintf("%v", rMessage)
		//Send message...
		util.SendNotifications(rUser.FCMToken, rMessageConv)
		return
	}
	return

}

// CallBackHandler is...
func CallBackHandler(w http.ResponseWriter, r *http.Request) {
	UpdatePayment(w, r)
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
	res.Error = err.Error()
	json.NewEncoder(w).Encode(res)
	return
}
//UpdatePayment
func UpdatePayment(w http.ResponseWriter, r *http.Request) {
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
	var phoneNumber interface{}
	var paymentModel model.Payment
	resultCodeString := fmt.Sprintf("%v", resultCode)
	resultDesc := fmt.Sprintf("%v", rBody)

	if resultCodeString == string('0') {
		item = bd.(map[string]interface{})["Body"].(map[string]interface{})["stkCallback"].(map[string]interface{})["CallbackMetadata"].(map[string]interface{})["Item"]
		mpesaReceiptNumber = item.([]interface{})[1].(map[string]interface{})["Value"]
		transactionDate = item.([]interface{})[3].(map[string]interface{})["Value"]
		phoneNumber = item.([]interface{})[4].(map[string]interface{})["Value"]

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
			"phoneNumber":        phoneNumber,
			"checkoutRequestID":  checkoutRequestID,
			"isSuccessful":       true,
		}}
		err = paymentCollection.FindOneAndUpdate(context.TODO(), paymentFilter, paymentUpdate).Decode(&paymentModel)
		if err != nil {
			fmt.Printf("error...")
			return

		}
		//Send message...
		util.SendNotifications(result.FCMToken, resultDesc)
		res.Result = "Payment updated"
		json.NewEncoder(w).Encode(res)
		return
	} else {
		paymentModel.IsSuccessful = false
		//Send message...
		util.SendNotifications(result.FCMToken, resultDesc)
		return
	}

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
		filter := bson.M{"vehicleReg": vehicleReg,"isSuccessful": true}
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
