package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"net/http"
	"vepa/model"
	"vepa/util"
)

// RegisterHandler is...
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	var user model.User
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &user)
	var res model.ResponseResult
	if err != nil {
		res.Error = err.Error()
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	collection, err := util.GetUserCollection()

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
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	var user model.User
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &user)
	if err != nil {
		log.Fatal(err)
	}

	collection, err := util.GetUserCollection()

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
			return nil, fmt.Errorf("Unexpected signing method.")
		}
		return []byte("secret"), nil
	})
	var params = mux.Vars(r)
	//Get id from parameters
	userid := params["id"]
	id, _ := primitive.ObjectIDFromHex(userid)
	var user model.User
	var res model.ResponseResult
	collection, err := util.GetUserCollection()
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
	collection, err := util.GetUserCollection()
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
	defer r.Body.Close()
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
	collection, err := util.GetUserCollection()
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


