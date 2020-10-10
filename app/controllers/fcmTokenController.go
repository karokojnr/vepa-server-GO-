package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/mgo.v2"
	model "vepa/app/models"
)

const FCMTokenCollection = "fcmtoken"

func SaveAttendantsFCM(c *gin.Context) {
	//db := *MongoConfig()
	db := c.MustGet("db").(*mgo.Database)
	fmt.Println("MONGO RUNNING...", db)

	fcmToken := model.FCMToken{}
	err := c.Bind(&fcmToken)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Getting Body",
		})
		return
	}
	filter := bson.M{}
	update := bson.M{"$set": bson.M{"fcmtoken": fcmToken.FCMToken}}
	err = db.C(FCMTokenCollection).Update(filter, update)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "Error Updating FCMToken",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Success Updating FCMToken",
		//"user":    &user,
	})

}
