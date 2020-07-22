package util

import (
	"fmt"
	"log"
	"github.com/appleboy/go-fcm"
)

func SendNotifications(fcmToken string, notificationBody string) {
	//Send message...
	msg := &fcm.Message{
		To: fcmToken,
		Data: map[string]interface{}{
			"title": "Vepa",
			"body":  &notificationBody,
		},
	}
	// Create a FCM client to send the message.
	// env.GoDotEnvVariable("FCM_SERVER_KEY")
	client, err := fcm.NewClient(GoDotEnvVariable("FCM_SERVER_KEY"))
	if err != nil {
		log.Fatalln(err)
	}
	// Send the message and receive the response without retries.
	response, err := client.Send(msg)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%#v\n", response)
	fmt.Println("notification sent...")
	return

}
