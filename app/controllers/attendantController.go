package controllers

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	model "vepa/app/models"
	"vepa/app/util"
)

func AttendantLoginHandler(c *gin.Context) {
	ctx := context.TODO()
	attendantCollection, err := util.GetCollection("attendants")

	if err != nil {
		util.SendError(c, "Cannot get attendant collection")
		return
	}
	attendant := model.Attendant{}
	err = c.Bind(&attendant)

	if err != nil {
		util.SendError(c, "Error Getting Body")
		return
	}
	var result model.Attendant
	err = attendantCollection.FindOne(ctx, bson.M{"email": attendant.Email}).Decode(&result)
	if err != nil {

		util.SendError(c, "Attendant account was not found")
		c.Abort()
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(attendant.Password))
	if err != nil {
		util.SendError(c, "Invalid password")
		c.Abort()
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
		util.SendError(c, "Error while generating token")
		c.Abort()
		return
	}
	result.Token = tokenString
	result.Password = ""
	result.Exp = exp
	c.JSON(200, gin.H{
		"message":   "Successful login",
		"attendant": &result,
	})
	return

}
func GetAttendantDetailsHandler(c *gin.Context) {

}
