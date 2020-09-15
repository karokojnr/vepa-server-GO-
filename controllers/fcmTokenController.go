package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
	"vepa/model"
	"vepa/util"
)

func SaveAttendantsFCM(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var fcmModel model.FCMToken
	var res model.ResponseResult
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &fcmModel)
	fcmCollection, err := util.GetCollection("fcmtoken")
	if err != nil {
		log.Println(err)
	}
	update := bson.M{"$set": bson.M{"fcmtoken": fcmModel.FCMToken}}
	_, err = fcmCollection.UpdateOne(context.TODO(), bson.M{}, update)
	if err != nil {
		fmt.Printf("error...")
		return

	}
	//_, err = fcmCollection.InsertOne(context.TODO(), fcmModel)
	//if err != nil {
	//	log.Println(err)
	//}
	util.Log("FCMToken Added Successfully")
	res.Result = "FCMToken Added Successfully"
	json.NewEncoder(w).Encode(res)
}
