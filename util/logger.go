package util

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"log"
	"os"
	"time"
)

func InitLogger() {
	logFolder := os.Getenv("LOG_FOLDER")
	appName := os.Getenv("APP_NAME")
	pwd, err := os.Getwd()
	logFile := fmt.Sprintf("%s/%s/%s-%s.log", pwd, logFolder, appName, "%Y-%m-%d")
	logFileLink := fmt.Sprintf("%s/%s/%s.log", pwd, logFolder, appName)
	writer, err := rotatelogs.New(
		logFile,
		rotatelogs.WithLinkName(logFileLink),
		rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithRotationCount(10000),
	)
	if err != nil {
		fmt.Println("Failed to initialize log file ", err.Error())
	}
	log.SetOutput(writer)
	return
}

/*
	Call this function to log an event, i.e util.Log("Something happened")
*/
func Log(msg ...interface{}) {
	fmt.Println(msg...)
	msg = append(msg, "\n----------------------------")
	log.Println(msg...)
}

