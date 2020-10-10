package controllers

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	model "vepa/app/models"
	"vepa/app/util"
)

func GetDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
	})

}
func GetAttendants(c *gin.Context) {
	var results []*model.Attendant
	ctx := context.TODO()
	attendantCollection, err := util.GetCollection("attendants")
	if err != nil {
		util.SendError(c, "Cannot get attendant collection")
		return
	}
	cur, err := attendantCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		var elem model.Attendant
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
	numOfAttendants := len(results)
	c.HTML(http.StatusOK, "manage-attendants.html", gin.H{
		"title":      "Attendants",
		"attendants": &results,
		"len": numOfAttendants,
	})
	return

}
func GetAddAttendant(c *gin.Context) {
	c.HTML(http.StatusOK, "add-attendant.html", gin.H{
	})

}
func PostAddAttendant(c *gin.Context) {
	ctx := context.TODO()
	attendantCollection, err := util.GetCollection("attendants")
	if err != nil {
		util.SendError(c, "Cannot get attendant collection")
		return
	}
	attendant := model.Attendant{
		ID:          primitive.NewObjectID(),
		Firstname:   c.PostForm("firstName"),
		Lastname:    c.PostForm("lastName"),
		Email:       c.PostForm("email"),
		IDNumber:    c.PostForm("idNumber"),
		PhoneNumber: c.PostForm("phoneNumber"),
		Password:    c.PostForm("password"),
	}
	err = c.Bind(&attendant)
	if err != nil {
		util.SendError(c, "Error Getting Body")
		return
	}
	var result model.Attendant
	err = attendantCollection.FindOne(ctx, bson.M{"email": attendant.Email}).Decode(&result)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			hash, err := bcrypt.GenerateFromPassword([]byte(attendant.Password), 5)
			if err != nil {
				util.SendError(c, "Error While Hashing Password, Try Again Later")
				return
			}
			attendant.Password = string(hash)
			attendant.ID = primitive.NewObjectID()
			_, err = attendantCollection.InsertOne(ctx, &attendant)

			if err != nil {
				util.SendError(c, "Error Inserting Attendant")
				c.Abort()
				return
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"id":    attendant.ID,
				"email": attendant.Email,
			})

			tokenString, err := token.SignedString([]byte("secret"))
			exp := 60 * 60
			fmt.Println("Expires in ... seconds: ")
			fmt.Println(exp)
			if err != nil {
				util.SendError(c, "Error while generating token,Try again")
				c.Abort()
				return
			}

			attendant.Token = tokenString
			attendant.Password = ""
			attendant.Exp = exp
			c.Redirect(http.StatusMovedPermanently, "/dashboard")
			return
		}
	}
	util.SendError(c, "Attendant already exists!!!")
	c.Abort()
	return
}
