package controllers

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/contrib/sessions"

	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	model "vepa/app/models"
	"vepa/app/util"
)

func RegisterAdminHandler(c *gin.Context) {
	ctx := context.TODO()
	adminCollection, err := util.GetCollection("admins")
	if err != nil {
		util.SendError(c, "Cannot get admins collection")
		return
	}
	admin := model.Admin{}
	err = c.Bind(&admin)
	if err != nil {
		util.SendError(c, "Error Getting Body")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(admin.Password), 5)

	if err != nil {
		util.SendError(c, "Error While Hashing Password, Try Again Later")
		return
	}
	admin.Password = string(hash)
	admin.AdminID = primitive.NewObjectID()
	_, err = adminCollection.InsertOne(ctx, &admin)
	if err != nil {
		util.SendError(c, "Error Inserting Admin")
		c.Abort()
		return
	}
	c.JSON(200, gin.H{
		"message": "Success Inserting Admin",
		"admin":   &admin,
	})
	return

}
func AdminLoginHandler(c *gin.Context) {
	ctx := context.TODO()
	adminCollection, err := util.GetCollection("admins")

	if err != nil {
		util.SendError(c, "Cannot get admin collection")
		return
	}
	//admin := model.Admin{
	//	Email:    c.PostForm("email"),
	//	Password: c.PostForm("password"),
	//}
	//err = c.Bind(&admin)

	//if err != nil {
	//	util.SendError(c, "Error Getting Body")
	//	return
	//}
	var result model.Admin
	err = adminCollection.FindOne(ctx, bson.M{"email": c.PostForm("email")}).Decode(&result)
	if err != nil {

		util.SendError(c, "Admin account was not found")
		c.Abort()
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(c.PostForm("password")))
	if err != nil {
		util.SendError(c, "Invalid password")
		c.Abort()
		return
	}
	session := sessions.Default(c)
	session.Set("adminId", result.AdminID.Hex())
	session.Set("username", result.Email)
	expiryTime := util.SessionExpiry(2).Format(time.RFC3339)
	session.Set("expiry", expiryTime)
	util.Log("Session expiry time:", expiryTime)
	if err := session.Save(); err != nil {
		util.Log("Unable to save session:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error starting your session"})
		return
	}

	//c.JSON(200, gin.H{
	//	"message": "Successful login",
	//	"user":    &result,
	//})
	//c.HTML(http.StatusOK, "index.html", gin.H{

	//})
	c.Redirect(302, "/dashboard")

	return

}
func GetLogout(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("adminId")
	username := session.Get("username")
	if userId == nil || username == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	session.Delete("adminId")
	session.Delete("username")
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.Redirect(302, "/auth/getAdminLogin")
	//c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})

}
func GetAdminLoginHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title":      "VEPA - Login",
		"pageHeader": "Login",
	})

}
func GetDashboard(c *gin.Context) {
	//Num of Attendants
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

	//Num of Customers
	var userResults []*model.User
	userCtx := context.TODO()
	userCollection, err := util.GetCollection("users")
	if err != nil {
		util.SendError(c, "Cannot get users collection")
		return
	}
	userCur, err := userCollection.Find(userCtx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	for userCur.Next(context.TODO()) {
		var userElem model.User
		err := userCur.Decode(&userElem)
		if err != nil {
			log.Fatal(err)
		}
		userResults = append(userResults, &userElem)
	}
	if err := userCur.Err(); err != nil {
		util.SendError(c, "Error!")
		return
	}
	_ = userCur.Close(context.TODO())
	numOfUsers := len(userResults)

	//Num of Vehicles
	var vehicleResults []*model.Vehicle
	vehicleCtx := context.TODO()
	vehicleCollection, err := util.GetCollection("vehicles")
	if err != nil {
		util.SendError(c, "Cannot get vehilces collection")
		return
	}
	vehicleCur, err := vehicleCollection.Find(vehicleCtx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	for vehicleCur.Next(context.TODO()) {
		var vehicleElem model.Vehicle
		err := vehicleCur.Decode(&vehicleElem)
		if err != nil {
			log.Fatal(err)
		}
		vehicleResults = append(vehicleResults, &vehicleElem)
	}
	if err := vehicleCur.Err(); err != nil {
		util.SendError(c, "Error!")
		return
	}
	_ = vehicleCur.Close(context.TODO())
	numOfVehicles := len(vehicleResults)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"numOfCustomers":  numOfUsers,
		"numOfAttendants": numOfAttendants,
		"numOfVehicles":   numOfVehicles,
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
		"len":        numOfAttendants,
	})
	return

}
func GetAddAttendant(c *gin.Context) {
	c.HTML(http.StatusOK, "add-attendant.html", gin.H{})

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
func GetCustomers(c *gin.Context) {
	var results []*model.User
	ctx := context.TODO()
	userCollection, err := util.GetCollection("users")
	if err != nil {
		util.SendError(c, "Cannot get users collection")
		return
	}
	cur, err := userCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		var elem model.User
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
	numOfUsers := len(results)
	c.HTML(http.StatusOK, "manage-customers.html", gin.H{
		"title":     "Users",
		"customers": &results,
		"len":       numOfUsers,
	})
	return

}

//func GetAddAttendant(c *gin.Context) {
//	c.HTML(http.StatusOK, "add-attendant.html", gin.H{
//	})
//
//}
