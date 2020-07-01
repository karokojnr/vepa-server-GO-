package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	// "reflect"
	"vepa/model"
	"vepa/util/db"
	"vepa/util/env"

	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	"github.com/appleboy/go-fcm"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// RegisterHandler is...
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user model.User
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &user)
	var res model.ResponseResult
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

	collection, err := db.GetUserCollection()

	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	var result model.User
	err = collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&result)

	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)

			if err != nil {
				res.Error = "Error While Hashing Password, Try Again Later"
				json.NewEncoder(w).Encode(res)
				return
			}
			user.Password = string(hash)
			user.ID = primitive.NewObjectID()

			_, err = collection.InsertOne(context.TODO(), user)
			if err != nil {
				res.Error = "Error While Creating User, Try Again"
				json.NewEncoder(w).Encode(res)
				return
			}

			if err != nil {
				res.Error = "Invalid password"
				json.NewEncoder(w).Encode(res)
				return
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"id":    result.ID,
				"email": result.Email,
			})

			tokenString, err := token.SignedString([]byte("secret"))
			exp := 60 * 60
			fmt.Println("Expires in ... seconds: ")
			fmt.Println(exp)
			if err != nil {
				res.Error = "Error while generating token,Try again"
				json.NewEncoder(w).Encode(res)
				return
			}

			result.Token = tokenString
			result.Password = ""
			result.Exp = exp

			res.Result = "Registration Successful"
			json.NewEncoder(w).Encode(result)
			return
		}

		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

	res.Result = "Username already Exists!!"
	json.NewEncoder(w).Encode(res)
	return
}

// LoginHandler is...
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user model.User
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &user)
	if err != nil {
		log.Fatal(err)
	}

	collection, err := db.GetUserCollection()

	if err != nil {
		log.Fatal(err)
	}
	var result model.User
	var res model.ResponseResult

	err = collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&result)

	if err != nil {
		res.Error = "Invalid username"
		json.NewEncoder(w).Encode(res)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))

	if err != nil {
		res.Error = "Invalid password"
		json.NewEncoder(w).Encode(res)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    result.ID,
		"email": result.Email,
	})

	tokenString, err := token.SignedString([]byte("secret"))
	exp := 60 * 60
	fmt.Println("Expires in ... seconds: ")
	fmt.Println(exp)
	if err != nil {
		res.Error = "Error while generating token,Try again"
		json.NewEncoder(w).Encode(res)
		return
	}

	result.Token = tokenString
	result.Password = ""
	result.Exp = exp

	json.NewEncoder(w).Encode(result)

}

// ProfileHandler is...
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("secret"), nil
	})
	var params = mux.Vars(r)
	//Get id from parameters
	userid := params["id"]
	id, _ := primitive.ObjectIDFromHex(userid)
	var user model.User
	var res model.ResponseResult
	collection, err := db.GetUserCollection()
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		filter := bson.M{"_id": id}
		err := collection.FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			res.Error = "Unsuccessful!"
			json.NewEncoder(w).Encode(res)
			return
		}
		json.NewEncoder(w).Encode(user)
		return
	}
	res.Error = err.Error()
	json.NewEncoder(w).Encode(res)
	return
}

// EditProfileHandler is...
func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("secret"), nil
	})
	// var params = mux.Vars(r)
	// //Get id from parameters
	// vehicleid := params["id"]
	// id, _ := primitive.ObjectIDFromHex(vehicleid)
	var user model.User
	var res model.ResponseResult
	// body, _ := ioutil.ReadAll(r.Body)
	// err = json.Unmarshal(body, &user)
	collection, err := db.GetUserCollection()
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["id"].(string)
		userID, _ := primitive.ObjectIDFromHex(id)
		filter := bson.M{"_id": userID}
		// fmt.Println(userID)
		// fmt.Println(filter)
		// Read update model from body request
		_ = json.NewDecoder(r.Body).Decode(&user)
		update := bson.M{"$set": bson.M{
			"firstName":   user.Firstname,
			"lastName":    user.Lastname,
			"email":       user.Email,
			"idNumber":    user.IDNumber,
			"phoneNumber": user.PhoneNumber,
		}}
		// fmt.Println(update)
		var result model.User
		err := collection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&result)
		if err != nil {
			res.Error = "Unsuccessful!"
			json.NewEncoder(w).Encode(res)

			return

		}
		// fmt.Println("Past error")
		user.ID = userID
		// res.Result = "User updated Successfully"
		json.NewEncoder(w).Encode(user)
		return
	}
}

// FCMTokenHandler is...
func FCMTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("secret"), nil
	})
	var user model.User
	var res model.ResponseResult
	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &user)
	collection, err := db.GetUserCollection()
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["id"].(string)
		userID, _ := primitive.ObjectIDFromHex(id)
		filter := bson.M{"_id": userID}
		update := bson.M{"$set": bson.M{"fcmtoken": user.FCMToken}}
		_, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			fmt.Printf("error...")
			return

		}
		res.Result = "FCMToken updated"
		json.NewEncoder(w).Encode(res)
		return
	}
}

// AddVehicleHandler is...
func AddVehicleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-TYpe", "application/json")
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("secret"), nil
	})
	var vehicle model.Vehicle

	var res model.ResponseResult
	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &vehicle)
	collection, err := db.GetVehicleCollection()
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	var result model.Vehicle
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		vehicle.UserID = claims["id"].(string)
		vehicle.VeicleID = primitive.NewObjectID()
		err = collection.FindOne(context.TODO(), bson.M{"registrationNumber": vehicle.RegistrationNumber}).Decode(&result)
		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				_, err = collection.InsertOne(context.TODO(), vehicle)
				if err != nil {
					res.Error = "Error While Adding Vehicle, Try Again"
					json.NewEncoder(w).Encode(res)
					return
				}
				res.Result = "Vehicle Added Successfully"
				json.NewEncoder(w).Encode(res)
				return
			}
		}
		res.Result = "Vehicle already Exists!!"
		json.NewEncoder(w).Encode(res)
		return
	}
	res.Error = err.Error()
	json.NewEncoder(w).Encode(res)
	return
}

// EditVehicleHandler is...
func EditVehicleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("secret"), nil
	})
	var params = mux.Vars(r)
	//Get id from parameters
	vehicleid := params["id"]
	id, _ := primitive.ObjectIDFromHex(vehicleid)
	var vehicle model.Vehicle
	var res model.ResponseResult
	// body, _ := ioutil.ReadAll(r.Body)
	// err = json.Unmarshal(body, &vehicle)
	collection, err := db.GetVehicleCollection()
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		//Use user Id
		// _ = claims["id"].(string)
		filter := bson.M{"_id": id}
		// Read update model from body request
		_ = json.NewDecoder(r.Body).Decode(&vehicle)
		update := bson.M{"$set": bson.M{
			"registrationNumber": vehicle.RegistrationNumber,
			"vehicleClass":       vehicle.VehicleClass,
		}}
		var result model.Vehicle
		err := collection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&result)
		if err != nil {
			res.Error = "Unsuccessful!"
			json.NewEncoder(w).Encode(res)

			return
		}
		vehicle.VeicleID = id

		json.NewEncoder(w).Encode(vehicle)
		// log.Println("Past could not update!")
		// res.Result = "Vehicle updated successfully"
		// // json.NewEncoder(w).Encode(res)
		// json.NewEncoder(w).Encode(res)
		return
	}
}

// UserVehiclesHandler is...
func UserVehiclesHandler(w http.ResponseWriter, r *http.Request) {
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
	var results []*model.Vehicle
	collection, err := db.GetVehicleCollection()
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
			var elem model.Vehicle
			err := cur.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}
			results = append(results, &elem)
		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}
		cur.Close(context.TODO())
		json.NewEncoder(w).Encode(results)
		return
	}
	res.Error = err.Error()
	json.NewEncoder(w).Encode(res)
	return
}

// PaymentHandler is...
func PaymentHandler(w http.ResponseWriter, r *http.Request) {
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
	collection, err := db.GetPaymentCollection()
	if err != nil {
		res.Error = "Error, Try Again Later"
		json.NewEncoder(w).Encode(res)
		return
	}
	var rUser model.User
	var result model.Payment
	err = collection.FindOne(context.TODO(), bson.M{"days": payment.Days, "isSuccessful": true}).Decode(&result)

	if err != nil {
		if err.Error() == "mongo: no documents in result" {
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
					appKey    = env.GoDotEnvVariable("MPESA_APP_KEY")
					appSecret = env.GoDotEnvVariable("MPESA_APP_SECRET")
				)
				svc, err := mpesa.New(appKey, appSecret, mpesa.SANDBOX)
				if err != nil {
					panic(err)
				}

				mres, err := svc.Simulation(mpesa.Express{
					BusinessShortCode: "174379",
					Password:          env.GoDotEnvVariable("MPESA_PASSWORD"),
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

					err = collection.FindOne(context.TODO(), filter).Decode(&rUser)
					if err != nil {
						if err.Error() == "mongo: no documents in result" {
							res.Result = "Something went wrong, Please try again later!"
							json.NewEncoder(w).Encode(res)
							return
						}
					}
					msg := &fcm.Message{
						To: rUser.FCMToken,
						Data: map[string]interface{}{
							"title": "Vepa",
							"body":  rMessage,
						},
					}
					// Create a FCM client to send the message.
					client, err := fcm.NewClient(env.GoDotEnvVariable("FCM_SERVER_KEY"))
					if err != nil {
						log.Fatalln(err)
					}
					// Send the message and receive the response without retries.
					response, err := client.Send(msg)
					if err != nil {
						log.Fatalln(err)
					}
					log.Printf("%#v\n", response)
					return
				}

				log.Println(mres)
				return
			}
		}
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	//Send message...
	msg := &fcm.Message{
		To: rUser.FCMToken,
		Data: map[string]interface{}{
			"title": "Vepa",
			"body":  "Kindly choose a date that you have not paid for!",
		},
	}
	// Create a FCM client to send the message.
	client, err := fcm.NewClient(env.GoDotEnvVariable("FCM_SERVER_KEY"))
	if err != nil {
		log.Fatalln(err)
	}
	// Send the message and receive the response without retries.
	response, err := client.Send(msg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%#v\n", response)
	fmt.Println("Already Paid")
	res.Result = "Payment already Exists!!"
	json.NewEncoder(w).Encode(res)
	return

	// res.Error = err.Error()
	// json.NewEncoder(w).Encode(res)
	// return

}

// CallBackHandler is...
func CallBackHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
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
	collection, err := db.GetUserCollection()
	if err != nil {
		log.Fatal(err)
	}
	//extract userId
	r.ParseForm() // Parses the request body
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

	paymentCollection, err := db.GetPaymentCollection()
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
	res.Result = "Payment updated"
	json.NewEncoder(w).Encode(res)

	//Send message...
	msg := &fcm.Message{
		To: result.FCMToken,
		Data: map[string]interface{}{
			"title": "Vepa",
			"body":  rBody,
		},
	}
	// Create a FCM client to send the message.
	client, err := fcm.NewClient(env.GoDotEnvVariable("FCM_SERVER_KEY"))
	if err != nil {
		log.Fatalln(err)
	}
	// Send the message and receive the response without retries.
	response, err := client.Send(msg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%#v\n", response)

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
	collection, err := db.GetPaymentCollection()
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
		cur.Close(context.TODO())
		json.NewEncoder(w).Encode(results)
		return
	}
	res.Error = err.Error()
	json.NewEncoder(w).Encode(res)
	return
}
