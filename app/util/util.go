package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"time"
)

func SendError(c *gin.Context, msg string) {
	Log(msg)
	c.JSON(http.StatusBadRequest, gin.H{
		"error": msg,
	})
}

func SendJson(c *gin.Context, payload gin.H) {
	c.JSON(http.StatusOK, payload)
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]interface{}{"error": msg})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	Log(fmt.Sprintf("RESPONSE:: Status:%d Payload: %v", code, payload))
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func ValidateEmail(email string) (bool, error) {
	var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if len(email) > 254 || !rxEmail.MatchString(email) {
		return false, errors.New("email is invalid")
	}
	return true, nil
}

func SessionExpiry(hours int) time.Time {
	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
	date = date.Add(time.Hour * time.Duration(hours))
	return date
}
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		Log("Hit an auth required endpoint...")
		session := sessions.Default(c)
		adminId := session.Get("adminId")
		username := session.Get("username")
		expiry := fmt.Sprintf("%v", session.Get("expiry"))
		if username == nil || adminId == nil {
			c.Redirect(302, "/auth/getAdminLogin")
			return
		}
		Log("Expiry:", expiry)
		expiryTime, err := time.Parse(time.RFC3339, expiry)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
		if time.Now().After(expiryTime) {
			//c.AbortWithStatusJSON(http.StatusBadRequest, "Session expired")
			c.Redirect(302, "/auth/getAdminLogin")
			return
		} else {
			expiryTime := SessionExpiry(2).Format(time.RFC3339)
			session.Set("expiry", expiryTime)
			session.Save()
		}

		//url := c.Request.URL.Path
		status := struct {
			Status int    `json:"status"`
			Value  string `json:"value"`
		}{}
		//sessionCollection, err := GetCollection("sessions")

		//if err != nil {
		//	SendError(c, "Cannot get admin collection")
		//	return
		//}
		//sessions.
		//query := `select up.status as status, p.value as value
		//from user_permission up
		//inner join permissions p on up.permission_id = p.id
		//where up.user_id=? and p.value = ? and up.deleted_at is null;`
		//app.DB.Raw(query, adminId, url).Scan(&status)

		if len(status.Value) != 0 {
			if status.Status == 1 {
				fmt.Println("Allowed")
			} else {
				fmt.Println("Content-Type", c.Request.Header)
				if c.Request.Header.Get("Content-Type") != "application/json;charset=utf-8" {
					fmt.Println("Not allowed")
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not authorized to access this route"})
					return
				}
			}
		}
		Log("Request made by admin Id:", adminId, "Username:", username)
		c.Set("userId", adminId)
		c.Set("username", username)
		c.Next()
	}
}
