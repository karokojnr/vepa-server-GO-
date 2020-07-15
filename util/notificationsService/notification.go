package notificationsService

import (
	"fmt"
	"github.com/appleboy/go-fcm"
	"log"
	"vepa/util/env"
)

func SendNotifcation(fcmToken string, notificationBody interface{})  {

	//Send message...
	msg := &fcm.Message{
		To: fcmToken,
		Data: map[string]interface{}{
			"title": "Vepa",
			"body":  notificationBody,
		},
	}
	fmt.Println("IM HERE...")
	// Create a FCM client to send the message.
	client, err := fcm.NewClient(env.GoDotEnvVariable("FCM_SERVER_KEY"))
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
