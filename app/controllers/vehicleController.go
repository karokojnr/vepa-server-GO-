package controllers

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	model "vepa/app/models"
	"vepa/app/util"
)

func AddVehicleHandler(c *gin.Context) {
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehicle collection")
		return
	}
	vehicle := model.Vehicle{}
	tokenString := c.GetHeader("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("secret"), nil
	})
	err = c.Bind(&vehicle)
	if err != nil {
		util.SendError(c, "Error getting body")
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		vehicle.UserID = claims["id"].(string)
		vehicle.VeicleID = primitive.NewObjectID()
		vehicle.IsWaitingClamp = false
		vehicle.IsClamped = false
		var result model.Vehicle
		err = vehicleCollection.FindOne(ctx, bson.M{"registrationNumber": vehicle.RegistrationNumber}).Decode(&result)
		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				_, err = vehicleCollection.InsertOne(ctx, vehicle)
				if err != nil {
					util.SendError(c, "Error Insert Vehicle")
					c.Abort()
					return
				}
				c.JSON(200, gin.H{
					"message": "Success Insert Vehicle",
					"vehicle": &vehicle,
				})
				return
			}
		}
		util.SendError(c, "Vehicle Already Exists")
		c.Abort()
		return
	}
	util.SendError(c, "You are not authorized!!!")
	c.Abort()
	return
}
func GetVehicleHandler(c *gin.Context) {
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehicle collection")
		return
	}

	vehicle := model.Vehicle{}
	vehicleReg := c.Param("vehicleReg")

	err = vehicleCollection.FindOne(ctx, bson.M{"registrationNumber": vehicleReg}).Decode(&vehicle)

	if err != nil {
		util.SendError(c, "Error Getting Vehicle")
		return
	}

	c.JSON(200, gin.H{
		"vehicle": &vehicle,
	})
	return
}
func EditVehicleHandler(c *gin.Context) {
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehicle collection")
		return
	}
	tokenString := c.GetHeader("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("secret"), nil
	})

	vehicle := model.Vehicle{}
	vehicleID := c.Param("id")
	id, _ := primitive.ObjectIDFromHex(vehicleID)
	err = c.Bind(&vehicle)

	if err != nil {
		util.SendError(c, "Error Getting Body")
		return
	}
	vehicle.VeicleID = id
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"registrationNumber": vehicle.RegistrationNumber,
		"vehicleClass":       vehicle.VehicleClass,
	}}
	var result model.Vehicle
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		err = vehicleCollection.FindOneAndUpdate(ctx, filter, update).Decode(&result)
		if err != nil {
			util.SendError(c, "Error Updating Vehicle")
			return
		}

		c.JSON(200, gin.H{
			"message": "Success Updating Vehicle",
			"vehicle": &vehicle,
		})
		return
	}
	util.SendError(c, "You are not authorized!!!")
	c.Abort()
	return
}
func UserVehiclesHandler(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("secret"), nil
	})
	var results []*model.Vehicle
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehicle collection")
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["id"].(string)
		filter := bson.M{"userId": userID}
		cur, err := vehicleCollection.Find(ctx, filter)
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
			util.SendError(c, "Error!")
			return
		}
		_ = cur.Close(context.TODO())
		c.JSON(200, gin.H{
			"vehicles": &results,
		})
		return
	}
	util.SendError(c, "You are not authorized!!!")
	c.Abort()
	return
}
func DeleteVehicleHandler(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte("secret"), nil
	})
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehicle collection")
		return
	}
	vehicleID := c.Param("id") // Get Param
	id, _ := primitive.ObjectIDFromHex(vehicleID)
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		filter := bson.M{"_id": id}
		var result model.Vehicle
		err = vehicleCollection.FindOne(ctx, filter).Decode(&result)
		if err != nil {
			log.Println(err)
		}
		util.Log("Vehicle to be deleted found - ", result.RegistrationNumber)
		if result.IsClamped == false {
			deleteResult, err := vehicleCollection.DeleteOne(ctx, filter)
			if err != nil {
				fmt.Println(err)
				return
			}
			util.Log("Deleted Vehicle")
			c.JSON(200, gin.H{
				"delete result": deleteResult,
			})
			return
		}
		util.Log("Vehicle Clamped! Deletion not allowed")
		c.JSON(200, gin.H{
			"message": "clamped",
		})
		return
	}
	util.SendError(c, "You are not authorized!!!")
	c.Abort()
	return

}
func VehiclesWaitingClamp(c *gin.Context) {
	var vehicles []*model.Vehicle
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehicle collection")
		return
	}
	vehicleFilter := bson.M{
		"isWaitingClamp": true,
		//"isClamped":      false,
	}
	cur, err := vehicleCollection.Find(ctx, vehicleFilter)
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
	c.JSON(200, gin.H{
		"vehicles": vehicles,
	})
	return

}
func ClampedVehicle(c *gin.Context) {
	var vehicles []*model.Vehicle
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehicle collection")
		return
	}
	vehicleFilter := bson.M{
		"isWaitingClamp": false,
		"isClamped":      true,
	}
	cur, err := vehicleCollection.Find(ctx, vehicleFilter)
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
	c.JSON(200, gin.H{
		"vehicles": vehicles,
	})
	return
}
func CheckIfVehicleIsClampedHandler(c *gin.Context) {
	ctx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehicle collection")
		return
	}
	vehicleReg := c.Param("vehicleReg")
	var vehicleModel model.Vehicle
	vehicleFilter := bson.M{"registrationNumber": vehicleReg, "isClamped": true}
	err = vehicleCollection.FindOne(ctx, vehicleFilter).Decode(&vehicleModel)
	if err != nil {
		util.SendJson(c, gin.H{
			"vehicle": vehicleModel,
		})
		return
	}
	util.SendError(c, "clamped")
	return
}
