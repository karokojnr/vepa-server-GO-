package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"vepa/model"
	"vepa/util/db"

	"github.com/AndroidStudyOpenSource/mpesa-api-go"
	"github.com/appleboy/go-fcm"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
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

			_, err = collection.InsertOne(context.TODO(), user)
			if err != nil {
				res.Error = "Error While Creating User, Try Again"
				json.NewEncoder(w).Encode(res)
				return
			}
			res.Result = "Registration Successful"
			json.NewEncoder(w).Encode(res)
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
		"exp":   time.Now().Add(time.Hour * time.Duration(1)).Unix(),
		"iat":   time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte("secret"))

	// exp := time.Now().Add(time.Hour * time.Duration(1)).Unix()
	exp := time.Now().Add(time.Hour * 15).Unix()
	fmt.Println(time.Now())
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
	var result model.User
	var res model.ResponseResult
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		result.Email = claims["email"].(string)
		json.NewEncoder(w).Encode(result)
		return
	}
	res.Error = err.Error()
	json.NewEncoder(w).Encode(res)
	return
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
	var resultUser model.User
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["id"].(string)
		filter := bson.M{"_id": userID}
		update := bson.M{
			"$set": bson.M{"fcmtoken": user.FCMToken},
		}
		result := collection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&resultUser)
		if result != nil {
			fmt.Printf("FCMToken updated")
		}
		res.Result = "Something went wrong"
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
		fmt.Printf("Found multiple documents (array of pointers ): %+v\n", results)
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
	var res model.ResponseResult
	if err != nil {
		res.Error = "Error, Try Again Later"
		json.NewEncoder(w).Encode(res)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["id"].(string)
		fmt.Println(userID)
		const (
			appKey    = "WRnVsZ32lzmgQOVAoiANPAB9se2RYrB2"
			appSecret = "ixv4HzhalH1fL9ry"
		)
		svc, err := mpesa.New(appKey, appSecret, mpesa.SANDBOX)
		if err != nil {
			panic(err)
		}

		res, err := svc.Simulation(mpesa.Express{
			BusinessShortCode: "174379",
			Password:          "MTc0Mzc5YmZiMjc5ZjlhYTliZGJjZjE1OGU5N2RkNzFhNDY3Y2QyZTBjODkzMDU5YjEwZjc4ZTZiNzJhZGExZWQyYzkxOTIwMjAwNDIxMTc1NTU1",
			Timestamp:         "20200421175555",
			TransactionType:   "CustomerPayBillOnline",
			Amount:            1,
			PartyA:            "254799338805",
			PartyB:            "174379",
			PhoneNumber:       "254799338805",
			CallBackURL:       "https://vepa-server-go.herokuapp.com/rcb",
			AccountReference:  "Vepa",
			TransactionDesc:   "Vepa Payment",
		})

		if err != nil {
			log.Println(err)
		}
		log.Println(res)
	}
	res.Error = err.Error()
	json.NewEncoder(w).Encode(res)
	return

}

// CallBackHandler is...
func CallBackHandler(w http.ResponseWriter, r *http.Request) {
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
	if err != nil {
		res.Error = "Error, Try Again Later"
		json.NewEncoder(w).Encode(res)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["id"].(string)
		fmt.Println(userID)
		fmt.Println("-----------Received M-Pesa webhook-----------")
		// fmt.Println(req.body)
		// fmt.Println(JSON.stringify(req.body.Body.stkCallback.ResultDesc))
		fmt.Println("---------------------------------------------")
		// Create the message to be sent.
		var user model.User
		collection, err := db.GetUserCollection()
		if err != nil {
			log.Fatal(err)
		}
		err = collection.FindOne(context.TODO(), bson.M{"_id": user.ID}).Decode(&user)
		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				res.Result = "Something went wrong, Please try again later!"
				json.NewEncoder(w).Encode(res)
				return
			}
		}
		msg := &fcm.Message{
			To: user.FCMToken,
			Data: map[string]interface{}{
				"foo": "bar",
			},
		}
		// Create a FCM client to send the message.
		client, err := fcm.NewClient("AIzaSyABBYrj6YeQxxqbgIsaouXJONZZ5Ecw2Sk")
		if err != nil {
			log.Fatalln(err)
		}

		// Send the message and receive the response without retries.
		response, err := client.Send(msg)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("%#v\n", response)

	}
}
