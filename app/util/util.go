package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
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
