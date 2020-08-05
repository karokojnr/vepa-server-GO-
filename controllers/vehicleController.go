package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io/ioutil"
	"log"
	"net/http"
	"vepa/model"
	"vepa/util"
)

// AddVehicleHandler is...
func AddVehicleHandler(w http.ResponseWriter, r *http.Request) {
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
	var vehicle model.Vehicle

	var res model.ResponseResult
	body, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &vehicle)
	collection, err := util.GetVehicleCollection()
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
	collection, err := util.GetVehicleCollection()
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
	collection, err := util.GetVehicleCollection()
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
		_ = cur.Close(context.TODO())
		json.NewEncoder(w).Encode(results)
		return
	}
	res.Error = err.Error()
	json.NewEncoder(w).Encode(res)
	return
}

//DeleteVehicleHandler is...
func DeleteVehicleHandler(w http.ResponseWriter, r *http.Request) {
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
	vehicleCollection, err := util.GetVehicleCollection()
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// prepare filter.
		filter := bson.M{"_id": id}

		deleteResult, err := vehicleCollection.DeleteOne(context.TODO(), filter)

		if err != nil {
			fmt.Println(err)
			return
		}

		json.NewEncoder(w).Encode(deleteResult)

	}

}
