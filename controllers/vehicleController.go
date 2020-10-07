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
	collection, err := util.GetCollection("vehicles")
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	var result model.Vehicle
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		vehicle.UserID = claims["id"].(string)
		vehicle.VeicleID = primitive.NewObjectID()
		vehicle.IsWaitingClamp = false
		vehicle.IsClamped = false
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
func GetVehicleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(r)
	//Get id from parameters
	vehicleReg := params["vehicleReg"]
	//id, _ := primitive.ObjectIDFromHex(vehicleid)
	var vehicleModel model.Vehicle
	var res model.ResponseResult
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}
	err = vehicleCollection.FindOne(context.TODO(), bson.M{"registrationNumber": vehicleReg}).Decode(&vehicleModel)
	if err != nil {
		log.Println(err)
	}
	json.NewEncoder(w).Encode(vehicleModel)
	return

}

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
	collection, err := util.GetCollection("vehicles")
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
	collection, err := util.GetCollection("vehicles")
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
		_ = cur.Close(context.TODO())
		json.NewEncoder(w).Encode(results)
		return
	}
	res.Error = "You are not Authorized"
	json.NewEncoder(w).Encode(res)
	return
}

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
	vehicleID := params["id"]
	var res model.ResponseResult
	id, _ := primitive.ObjectIDFromHex(vehicleID)
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// prepare filter.
		filter := bson.M{"_id": id}
		var result model.Vehicle
		err = vehicleCollection.FindOne(context.TODO(), filter).Decode(&result)
		if err != nil {
			log.Println(err)
		}
		util.Log("Vehicle to be deleted found - ", result.RegistrationNumber)
		if result.IsClamped == false {
			deleteResult, err := vehicleCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
				return
			}
			util.Log("Deleted Vehicle")
			json.NewEncoder(w).Encode(deleteResult)
			return
		}
		util.Log("Vehicle Clamped! Deletion not allowed")
		res.Error = "clamped"
		json.NewEncoder(w).Encode(res)
		return

	}

}
func VehiclesWaitingClamp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var vehicles []*model.Vehicle
	vehicleColection, err := util.GetCollection("vehicles")
	if err != nil {
		log.Println(err)
	}
	vehicleFilter := bson.M{
		"isWaitingClamp": true,
		//"isClamped":      false,
	}
	cur, err := vehicleColection.Find(context.TODO(), vehicleFilter)
	if err != nil {
		log.Println(err)
	}
	for cur.Next(context.TODO()) {
		var elem model.Vehicle
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		vehicles = append(vehicles, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	_ = cur.Close(context.TODO())
	json.NewEncoder(w).Encode(vehicles)
	return
}

func ClampedVehicles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var vehicles []*model.Vehicle
	vehicleColection, err := util.GetCollection("vehicles")
	if err != nil {
		log.Println(err)
	}
	vehicleFilter := bson.M{
		"isClamped":      true,
		"isWaitingClamp": false,
	}
	cur, err := vehicleColection.Find(context.TODO(), vehicleFilter)
	if err != nil {
		log.Println(err)
	}
	for cur.Next(context.TODO()) {
		var elem model.Vehicle
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		vehicles = append(vehicles, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	_ = cur.Close(context.TODO())
	json.NewEncoder(w).Encode(vehicles)
	return
}
