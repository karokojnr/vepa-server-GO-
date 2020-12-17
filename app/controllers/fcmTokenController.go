package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	model "vepa/app/models"
	"vepa/app/util"
)

func SaveAttendantsFCMHandler(c *gin.Context) {
	ctx := context.TODO()
	attendantCollection, err := util.GetCollection("attendants")

	if err != nil {
		util.SendError(c, "Cannot get attendant collection")
		return
	}

	attendant := model.User{}
	attendantID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(attendantID)
	err = c.Bind(&attendant)

	if err != nil {
		util.SendError(c, "Error Getting Body")
		return
	}
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"fcmtoken": attendant.FCMToken}}
	_, err = attendantCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		util.SendError(c, "Error Updating FCMToken")
		return
	}

	c.JSON(200, gin.H{
		"message": "Success Updating FCMToken",
		//"attendant":    &attendant,
	})
	return
}
